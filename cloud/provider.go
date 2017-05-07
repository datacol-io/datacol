package cloud

import (
	"io"

	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud/google"
)

type Provider interface {
	AppCreate(string) (*pb.App, error)
	AppGet(string) (*pb.App, error)
	AppDelete(string) error
	AppRestart(string) error
	AppList() (pb.Apps, error)

	BuildImport(key string, source []byte) error
	BuildCreate(app, source string, opts *pb.BuildOptions) error

	EnvironmentGet(app string) (pb.Environment, error)
	EnvironmentSet(app string, body io.Reader) error

	LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error

	BuildList(app string, limit int) (pb.Builds, error)
	BuildGet(app, id string) (*pb.Build, error)
	// BuildSave(*client.Build) error
	BuildDelete(app, id string) error
	// BuildLogs(app, id string, w io.Writer) error
	BuildRelease(*pb.Build) (*pb.Release, error)

	ReleaseList(string, int) (pb.Releases, error)
	ReleaseDelete(string, string) (error)

	ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error)
	ResourceDelete(name string) error
	ResourceGet(name string) (*pb.Resource, error)
	ResourceLink(app, name string) (*pb.Resource, error)
	ResourceList() (pb.Resources, error)
	ResourceUnlink(string, string) (*pb.Resource, error)
	// ResourceUpdate(name string, params map[string]string) (*pb.Resource, error)

	GetRunningPods(string) (string, error)
}

func Getgcp(name, project string) Provider {
	return &google.GCPCloud{
		Project:        project,
		DeploymentName: name,
	}
}
