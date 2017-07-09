package main

import (
	"cloud.google.com/go/compute/metadata"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/appscode/cloudid"
	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud"
	daws "github.com/dinesh/datacol/cloud/aws"
	"github.com/dinesh/datacol/cloud/google"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

func newServer() *Server {
	var password, name string
	var provider cloud.Provider

	cid := cloudid.Detect()
	switch cid {
	case "aws":
		password = os.Getenv("DATACOL_API_KEY")
		name = os.Getenv("DATACOL_STACK")
		region := os.Getenv("AWS_REGION")

		if len(name) == 0 || len(password) == 0 {
			log.Fatal("unable to find DATACOL_STACK or DATACOL_API_KEY env vars.")
		}

		if region == "" {
			log.Fatal("AWS_REGION env var not found")
		}

		provider = &daws.AwsCloud{
			DeploymentName: name,
			Region:         region,
			SettingBucket:  os.Getenv("DATACOL_BUCKET"),
		}
	case "gce":
		var bucket, zone, projectId, projectNumber string
		password = getAttr("DATACOL_API_KEY")
		bucket = getAttr("DATACOL_BUCKET")
		name = getAttr("DATACOL_STACK")
		z, err := metadata.Zone()
		if err != nil {
			log.Fatal(err)
		}
		zone = z

		projectId, err = metadata.ProjectID()
		if err != nil {
			log.Fatal(err)
		}

		projectNumber, err = metadata.NumericProjectID()
		if err != nil {
			log.Fatal(err)
		}

		provider = &google.GCPCloud{
			Project:        projectId,
			DeploymentName: name,
			BucketName:     bucket,
			Zone:           zone,
			ProjectNumber:  projectNumber,
		}

	default:
		log.Fatal(fmt.Errorf("Unsupported cloud: %s", cid))
	}

	return &Server{Provider: provider, Password: password, StackName: name}
}

func getAttr(key string) string {
	v, err := metadata.InstanceAttributeValue(key)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

type Server struct {
	cloud.Provider
	StackName string
	Password  string
}

func (s *Server) Run() error {
	if _, err := s.Provider.K8sConfigPath(); err != nil {
		log.Warn(fmt.Errorf("caching kubernetes config err: %v", err))
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		return fmt.Errorf("opening rpcPort err: %v", err)
	}

	//todo: setting the max size to be 50MB. Add streaming for code upload
	maxMsgSize := 1024 * 1024 * 50

	// https://github.com/grpc/grpc-go/issues/106
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
		grpc.MaxMsgSize(maxMsgSize),
	)

	pbs.RegisterProviderServiceServer(grpcServer, s)

	return grpcServer.Serve(listener)
}

func (s *Server) Auth(ctx context.Context, req *pbs.AuthRequest) (*pbs.AuthResponse, error) {
	if authorize(ctx, s.Password) {
		var ip, project string
		if metadata.OnGCE() {
			_ip, err := metadata.ExternalIP()
			if err != nil {
				return nil, internalError(err, "couldn't resolve external ip for instance.")
			}
			ip = _ip
			_pid, err := metadata.ProjectID()
			if err != nil {
				return nil, internalError(err, "couldn't get projectId from metadata server.")
			}
			project = _pid
		} else {
			ip = "localhost"
			project = "gcs-local"
		}

		return &pbs.AuthResponse{Name: s.StackName, Host: ip, Project: project}, nil
	} else {
		return nil, internalError(fmt.Errorf("Invalid login trial with %s", req.Password), "invalid password")
	}
}

func (s *Server) AppCreate(ctx context.Context, req *pbs.AppRequest) (*pb.App, error) {
	return s.Provider.AppCreate(req.Name)
}

func (s *Server) AppGet(ctx context.Context, req *pbs.AppRequest) (*pb.App, error) {
	return s.Provider.AppGet(req.Name)
}

func (s *Server) AppList(ctx context.Context, req *pbs.ListRequest) (*pbs.AppListResponse, error) {
	apps, err := s.Provider.AppList()
	if err != nil {
		return nil, internalError(err, "unable to get apps")
	}

	result := &pbs.AppListResponse{Limit: req.Limit, Offset: req.Offset}
	result.Apps = apps
	return result, nil
}

func (s *Server) AppDelete(ctx context.Context, req *pbs.AppRequest) (*empty.Empty, error) {
	if err := s.Provider.AppDelete(req.Name); err != nil {
		return nil, internalError(err, "unable to delete app")
	}

	return &empty.Empty{}, nil
}

func (s *Server) AppRestart(ctx context.Context, req *pbs.AppRequest) (*empty.Empty, error) {
	if err := s.Provider.AppRestart(req.Name); err != nil {
		return nil, internalError(err, "unable to restart app")
	}

	return &empty.Empty{}, nil
}

func (s *Server) BuildCreate(stream pbs.ProviderService_BuildCreateServer) error {
	var app string
	fd, err := ioutil.TempFile(os.TempDir(), "upload-")
	if err != nil {
		return err
	}
	defer os.Remove(fd.Name())

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err = writeToFd(fd, req.Data); err != nil {
			return err
		}

		app = req.App
	}

	b, err := s.Provider.BuildCreate(app, fd.Name())
	if err != nil {
		return internalError(err, "failed to upload source.")
	}

	return stream.SendAndClose(b)
}

func (s *Server) BuildList(ctx context.Context, req *pbs.AppRequest) (*pbs.BuildListResponse, error) {
	items, err := s.Provider.BuildList(req.Name, 100)
	if err != nil {
		return nil, err
	}
	return &pbs.BuildListResponse{Builds: items}, nil
}

func (s *Server) BuildGet(ctx context.Context, req *pbs.AppIdRequest) (*pb.Build, error) {
	b, err := s.Provider.BuildGet(req.App, req.Id)
	if err != nil {
		return nil, internalError(err, "failed to get build.")
	}
	return b, nil
}

func (s *Server) BuildDelete(ctx context.Context, req *pbs.AppIdRequest) (*empty.Empty, error) {
	if err := s.Provider.BuildDelete(req.App, req.Id); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Server) BuildRelease(ctx context.Context, b *pb.Build) (*pb.Release, error) {
	r, err := s.Provider.BuildRelease(b)
	if err != nil {
		return nil, internalError(err, "failed to deploy app.")
	}
	return r, nil
}

func (s *Server) BuildLogs(ctx context.Context, req *pbs.BuildLogRequest) (*pbs.BuildLogResponse, error) {
	pos, lines, err := s.Provider.BuildLogs(req.App, req.Id, int(req.Pos))
	if err != nil {
		return nil, internalError(err, "build process failed.")
	}

	return &pbs.BuildLogResponse{Pos: int32(pos), Lines: lines}, nil
}

func (s *Server) BuildLogsStream(req *pbs.BuildLogStreamReq, stream pbs.ProviderService_BuildLogsStreamServer) error {
	reader, err := s.Provider.BuildLogsStream(req.Id)

	if err != nil {
		return err
	}

	buf := make([]byte, 0, 4*1024)

	for {
		n, err := reader.Read(buf[:cap(buf)])
		if err != nil {
			if n == 0 || err == io.EOF {
				return nil
			}
			return err
		}
		buf = buf[:n]

		if err := stream.Send(&pbs.LogStreamResponse{Data: buf}); err != nil {
			return err
		}
	}
}

// Releases endpoints
func (s *Server) ReleaseList(ctx context.Context, req *pbs.AppRequest) (*pbs.ReleaseListResponse, error) {
	items, err := s.Provider.ReleaseList(req.Name, 20)
	if err != nil {
		return nil, internalError(err, "failed to deploy app.")
	}
	return &pbs.ReleaseListResponse{Releases: items}, nil
}

func (s *Server) ReleaseDelete(ctx context.Context, req *pbs.AppIdRequest) (*empty.Empty, error) {
	err := s.Provider.ReleaseDelete(req.App, req.Id)
	if err != nil {
		return nil, internalError(err, "failed to delete release.")
	}
	return &empty.Empty{}, nil
}

func (s *Server) EnvironmentGet(ctx context.Context, req *pbs.AppRequest) (*pb.EnvConfig, error) {
	env, err := s.Provider.EnvironmentGet(req.Name)
	if err != nil {
		return nil, internalError(err, "failed to fetch env.")
	}
	return &pb.EnvConfig{Data: env}, nil
}

func (s *Server) EnvironmentSet(ctx context.Context, req *pbs.EnvSetRequest) (*empty.Empty, error) {
	err := s.Provider.EnvironmentSet(req.Name, strings.NewReader(req.Data))
	if err != nil {
		return nil, internalError(err, "failed to set env.")
	}
	return &empty.Empty{}, nil
}

func (s *Server) ResourceList(ctx context.Context, req *pbs.ListRequest) (*pbs.ResourceListResponse, error) {
	ret, err := s.Provider.ResourceList()
	if err != nil {
		return nil, err
	}
	return &pbs.ResourceListResponse{Resources: ret}, nil
}

func (s *Server) ResourceCreate(ctx context.Context, req *pbs.CreateResourceRequest) (*pb.Resource, error) {
	return s.Provider.ResourceCreate(req.Name, req.Kind, req.Params)
}

func (s *Server) ResourceGet(ctx context.Context, req *pbs.AppRequest) (*pb.Resource, error) {
	return s.Provider.ResourceGet(req.Name)
}

func (s *Server) ResourceDelete(ctx context.Context, req *pbs.AppRequest) (*empty.Empty, error) {
	if err := s.Provider.ResourceDelete(req.Name); err != nil {
		return nil, internalError(err, fmt.Sprintf("could not delete %s", req.Name))
	}
	return &empty.Empty{}, nil
}

func (s *Server) ResourceLink(ctx context.Context, req *pbs.AppResourceReq) (*pb.Resource, error) {
	ret, err := s.Provider.ResourceLink(req.App, req.Name)
	if err != nil {
		return nil, internalError(err, fmt.Sprintf("failed to link resource %s", req.Name))
	}
	return ret, nil
}

func (s *Server) ResourceUnlink(ctx context.Context, req *pbs.AppResourceReq) (*pb.Resource, error) {
	ret, err := s.Provider.ResourceUnlink(req.App, req.Name)
	if err != nil {
		return nil, internalError(err, fmt.Sprintf("failed to unlink resource %s", req.Name))
	}
	return ret, nil
}

func (s *Server) LogStream(req *pbs.LogStreamReq, stream pbs.ProviderService_LogStreamServer) error {
	since, err := ptypes.Duration(req.Since)
	if err != nil {
		return err
	}

	reader, close, err := s.Provider.LogStream(req.Name, pb.LogStreamOptions{
		Follow: req.Follow,
		Since:  since,
	})

	if err != nil {
		return err
	}
	defer close()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if err := stream.Send(&pbs.LogStreamResponse{Data: line}); err != nil {
			return err
		}
	}
}

func (s *Server) unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if !authorize(ctx, s.Password) {
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication required")
	}

	return handler(ctx, req)
}
