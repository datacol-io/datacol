package cloud

import (
	"bufio"
	pb "github.com/dinesh/datacol/api/models"
	"io"
)

type Provider interface {
	AppCreate(string) (*pb.App, error)
	AppGet(string) (*pb.App, error)
	AppDelete(string) error
	AppRestart(string) error
	AppList() (pb.Apps, error)

	EnvironmentGet(app string) (pb.Environment, error)
	EnvironmentSet(app string, body io.Reader) error

	BuildCreate(app string, data []byte) (*pb.Build, error)
	BuildList(app string, limit int) (pb.Builds, error)
	BuildGet(app, id string) (*pb.Build, error)
	BuildDelete(app, id string) error
	BuildRelease(*pb.Build) (*pb.Release, error)
	BuildLogs(app, id string, index int) (int, []string, error)

	ReleaseList(string, int) (pb.Releases, error)
	ReleaseDelete(string, string) error

	LogStream(app string, opts pb.LogStreamOptions) (*bufio.Reader, func() error, error)

	ResourceList() (pb.Resources, error)
	ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error)
	ResourceGet(name string) (*pb.Resource, error)
	ResourceDelete(name string) error
	ResourceLink(app, name string) (*pb.Resource, error)
	ResourceUnlink(string, string) (*pb.Resource, error)

	GetRunningPods(string) (string, error)
	K8sConfigPath() (string, error)
}
