package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"

	gcp_metadata "cloud.google.com/go/compute/metadata"
	log "github.com/Sirupsen/logrus"
	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	aws_provider "github.com/datacol-io/datacol/cloud/aws"
	"github.com/datacol-io/datacol/cloud/google"
	"github.com/datacol-io/datacol/cloud/local"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var emptyMsg = &empty.Empty{}

func newServer() *Server {
	var provider cloud.Provider

	cid, ok := os.LookupEnv("DATACOL_PROVIDER")
	if !ok {
		log.Fatalf("Missing provider env var. Please set `DATACOL_PROVIDER` in your shell")
	}

	name := os.Getenv("DATACOL_STACK")
	password := os.Getenv("DATACOL_API_KEY")

	if name == "" {
		log.Fatal("Missing `DATACOL_STACK` environment variable")
	}

	switch cid {
	case "aws":
		region := os.Getenv("AWS_REGION")

		if len(name) == 0 || len(password) == 0 {
			log.Fatal("unable to find DATACOL_STACK or DATACOL_API_KEY env vars.")
		}

		if region == "" {
			log.Fatal("AWS_REGION env var not found")
		}

		awsProvider := &aws_provider.AwsCloud{
			DeploymentName: name,
			Region:         region,
			SettingBucket:  os.Getenv("DATACOL_BUCKET"),
		}

		awsProvider.Setup()
		provider = awsProvider
	case "gcp":
		var bucket, zone, projectId, projectNumber string
		bucket = os.Getenv("DATACOL_BUCKET")
		zone = os.Getenv("GCP_DEFAULT_ZONE")
		region := os.Getenv("GCP_REGION")
		projectId = os.Getenv("GCP_PROJECT")
		projectNumber = os.Getenv("GCP_PROJECT_NUMBER")

		gcpProvider := &google.GCPCloud{
			Project:        projectId,
			DeploymentName: name,
			BucketName:     bucket,
			DefaultZone:    zone,
			Region:         region,
			ProjectNumber:  projectNumber,
		}

		gcpProvider.Setup()

		provider = gcpProvider

	case "local":
		// Local provider uses registry inside minikube VM to store images. Execute following commands to the registry running
		// inside minikube vm
		//
		// minikube start --insecure-registry localhost:5000 && \
		// 	eval $(minikube docker-env) && \
		// 	docker run -d -p 5000:5000 --restart=always --name registry registry:2

		localProvider := &local.LocalCloud{
			Name:            name,
			RegistryAddress: "localhost:5000",
			EnvMap:          make(map[string]pb.Environment),
		}
		localProvider.Setup()

		provider = localProvider
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
	if _, err := s.Provider.K8sConfigPath(); err != nil {
		log.Warnf("attempting to cache kubeconfig: %v. ", err)
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
	provider := cloud.CloudProvider(os.Getenv("DATACOL_PROVIDER"))
	if authorize(ctx, s.Password) {
		var ip, project string
		switch provider {
		case cloud.GCPProvider:
			if gcp_metadata.OnGCE() {
				_pid, err := gcp_metadata.ProjectID()
				if err != nil {
					return nil, internalError(err, "couldn't get projectId from metadata server.")
				}
				project = _pid
			}
		case cloud.AwsProvider:
			project = ""
		case cloud.LocalProvider:
			ip = "localhost"
			project = "gcs-local"
		default:
			return nil, fmt.Errorf("Invalid cloud provider: %s", provider)
		}

		return &pbs.AuthResponse{
			Provider: string(provider),
			Name:     s.StackName,
			Host:     ip,
			Project:  project,
		}, nil
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

	return emptyMsg, nil
}

func (s *Server) AppRestart(ctx context.Context, req *pbs.AppRequest) (*empty.Empty, error) {
	if err := s.Provider.AppRestart(req.Name); err != nil {
		return nil, internalError(err, "unable to restart app")
	}

	return emptyMsg, nil
}

func (s *Server) BuildCreate(ctx context.Context, req *pbs.CreateBuildRequest) (*pb.Build, error) {
	return s.Provider.BuildCreate(req.App, &pb.CreateBuildOptions{
		Procfile:  req.Procfile,
		Version:   req.Version,
		Trigger:   req.Trigger,
		DockerTag: req.DockerTag,
	})
}

func (s *Server) BuildUpload(stream pbs.ProviderService_BuildUploadServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return internalError(fmt.Errorf("No context found"), "No context found")
	}
	buildId := md["id"][0]

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
	}

	if err := fd.Close(); err != nil {
		return internalError(err, "failed to close tmpfile")
	}

	if err := s.Provider.BuildUpload(buildId, fd.Name()); err != nil {
		return internalError(err, "failed to upload source.")
	}

	return stream.SendAndClose(emptyMsg)
}

func (s *Server) BuildList(ctx context.Context, req *pbs.AppListRequest) (*pbs.BuildListResponse, error) {
	items, err := s.Provider.BuildList(req.Name, req.Limit)
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
	return emptyMsg, nil
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

func (s *Server) BuildImport(ws *websocket.Conn) error {
	buildId := ws.Request().Header.Get("id")
	if buildId == "" {
		return errors.New("Missing required header: id")
	}

	return s.Provider.BuildImport(buildId, ws, ws)
}

// Streaming build logs with websocket
func (s *Server) BuildLogStreamReq(ws *websocket.Conn) error {
	buildId := ws.Request().Header.Get("id")
	if buildId == "" {
		return errors.New("Missing required header: id")
	}

	r, err := s.Provider.BuildLogsStream(buildId)
	if err != nil {
		return err
	}

	//FIXME: r can be nil for minikube-based environment. Should return io.Reader from docker daemon
	if r != nil {
		_, err = io.Copy(ws, r)
	}

	return err
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
	return emptyMsg, nil
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
	return emptyMsg, nil
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
	return emptyMsg, nil
}

func (s *Server) ResourceLink(ctx context.Context, req *pbs.AppResourceReq) (*pb.Resource, error) {
	ret, err := s.Provider.ResourceLink(req.App, req.Resource)
	if err != nil {
		return nil, internalError(err, fmt.Sprintf("failed to link resource %s", req.Resource))
	}
	return ret, nil
}

func (s *Server) ResourceUnlink(ctx context.Context, req *pbs.AppResourceReq) (*pb.Resource, error) {
	ret, err := s.Provider.ResourceUnlink(req.App, req.Resource)
	if err != nil {
		return nil, internalError(err, fmt.Sprintf("failed to unlink resource %s", req.Resource))
	}
	return ret, nil
}

func (s *Server) DockerCredsGet(ctx context.Context, _ *empty.Empty) (*pb.DockerCred, error) {
	return s.Provider.DockerCredsGet()
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
