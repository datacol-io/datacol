package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	log "github.com/Sirupsen/logrus"
	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud"
	aws_provider "github.com/dinesh/datacol/cloud/aws"
	"github.com/dinesh/datacol/cloud/google"
	"github.com/dinesh/datacol/cloud/local"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func newServer() *Server {
	var provider cloud.Provider

	cid, ok := os.LookupEnv("DATACOL_PROVIDER")
	if !ok {
		log.Fatalf("Missing provider env var. Please set `DATACOL_PROVIDER` in your shell")
	}

	name := os.Getenv("DATACOL_STACK")
	password := os.Getenv("DATACOL_API_KEY")

	switch cid {
	case "aws":
		region := os.Getenv("AWS_REGION")

		if len(name) == 0 || len(password) == 0 {
			log.Fatal("unable to find DATACOL_STACK or DATACOL_API_KEY env vars.")
		}

		if region == "" {
			log.Fatal("AWS_REGION env var not found")
		}

		provider = &aws_provider.AwsCloud{
			DeploymentName: name,
			Region:         region,
			SettingBucket:  os.Getenv("DATACOL_BUCKET"),
		}
	case "gcp":
		var bucket, zone, projectId, projectNumber string
		bucket = os.Getenv("DATACOL_BUCKET")
		zone = os.Getenv("GCP_DEFAULT_ZONE")
		region := os.Getenv("GCP_REGION")
		projectId = os.Getenv("GCP_PROJECT")
		projectNumber = os.Getenv("GCP_PROJECT_NUMBER")

		provider = &google.GCPCloud{
			Project:        projectId,
			DeploymentName: name,
			BucketName:     bucket,
			DefaultZone:    zone,
			Region:         region,
			ProjectNumber:  projectNumber,
		}

	case "local":
		provider = &local.LocalCloud{
			Name:            name,
			RegistryAddress: "localhost:5000",
			EnvMap:          make(map[string]pb.Environment),
		}
	default:
		log.Fatalf("Unsupported cloud provider: %s", cid)
	}

	return &Server{Provider: provider, Password: password, StackName: name}
}

type Server struct {
	cloud.Provider
	StackName string
	Password  string
}

func (s *Server) Run() error {
	//TODO: should we expose K8sConfigPath for a provider ?
	if _, err := s.Provider.K8sConfigPath(); err != nil {
		log.Warn(fmt.Errorf("caching kubernetes config err: %v", err))
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		return fmt.Errorf("opening rpc port: %v", err)
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
	return s.Provider.AppCreate(req.Name, &pb.AppCreateOptions{
		RepoUrl: req.RepoUrl,
	})
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

func (s *Server) BuildCreate(ctx context.Context, req *pbs.CreateBuildRequest) (*pb.Build, error) {
	return s.Provider.BuildCreate(req.App, &pb.CreateBuildOptions{
		Version: req.Version,
	})
}

func (s *Server) BuildImport(stream pbs.ProviderService_BuildImportServer) error {
	var app string
	fd, err := ioutil.TempFile(os.TempDir(), "upload-")
	if err != nil {
		return err
	}

	log.Debugf("storing upload into %s", fd.Name())
	defer os.Remove(fd.Name())

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		log.Debugf("writing %d data", len(req.Data))

		if err = writeToFd(fd, req.Data); err != nil {
			log.Error(err)
			return err
		}

		app = req.App
	}

	if err := fd.Close(); err != nil {
		return internalError(err, "failed to close tmpfile")
	}

	b, err := s.Provider.BuildImport(app, fd.Name())
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

func (s *Server) BuildRelease(ctx context.Context, req *pbs.CreateReleaseRequest) (*pb.Release, error) {
	r, err := s.Provider.BuildRelease(req.Build, pb.ReleaseOptions{
		Domain: req.Domain,
	})

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

	if err != nil || reader == nil {
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
