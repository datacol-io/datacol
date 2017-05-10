package main

import (
	"cloud.google.com/go/compute/metadata"
	"fmt"
	"strings"
	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud"
	"github.com/dinesh/datacol/cloud/google"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

func newServer() *Server {
	var password, bucket, name, zone, projectId, projectNumber string

	if metadata.OnGCE() {
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

	} else {
		password = "secret"
		bucket = "datacol-gcs-local"
		name = "aaa"
		zone = "us-east1-b"
		projectId = "gcs-local"
		projectNumber = ""
	}

	provider := &google.GCPCloud{
		Project:        projectId,
		DeploymentName: name,
		BucketName:     bucket,
		Zone:           zone,
		ProjectNumber:  projectNumber,
	}

	return &Server{Provider: provider, Password: password}
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
	Password, StackName string
}

func (s *Server) Auth(ctx context.Context, req *pbs.AuthRequest) (*pbs.AuthResponse, error) {
	if req.Password == s.Password {
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
			ip = "localhost:8080"
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

func (s *Server) BuildCreate(ctx context.Context, req *pbs.CreateBuildRequest) (*pb.Build, error) {
	b, err := s.Provider.BuildCreate(req.App, req.Data)
	if err != nil {
		return nil, internalError(err, "failed to upload source.")
	}
	return b, nil
}

func (s *Server) BuildGet(ctx context.Context, req *pbs.GetBuildRequest) (*pb.Build, error) {
  b, err := s.Provider.BuildGet(req.App, req.Id)
  if err != nil {
    return nil, internalError(err, "failed to get build.")
  }
  return b, nil
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

func (s *Server) ResourceList(ctx context.Context, req *empty.Empty) (*pbs.ResourceListResponse, error) {
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
		return nil, internalError(err, fmt.Sprintf("failed to link resource %s", req.Name))
	}
	return ret, nil
}

