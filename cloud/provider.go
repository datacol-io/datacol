package cloud

import (
	"io"

	pb "github.com/dinesh/datacol/api/models"
)

type Provider interface {
	AppCreate(name string, req *pb.AppCreateOptions) (*pb.App, error)
	AppGet(string) (*pb.App, error)
	AppDelete(string) error
	AppRestart(string) error
	AppList() (pb.Apps, error)

	EnvironmentGet(app string) (pb.Environment, error)
	EnvironmentSet(app string, body io.Reader) error

	BuildCreate(app string, req *pb.CreateBuildOptions) (*pb.Build, error)
	BuildImport(app, filename string) (*pb.Build, error)
	BuildList(app string, limit int) (pb.Builds, error)
	BuildGet(app, id string) (*pb.Build, error)
	BuildDelete(app, id string) error
	BuildRelease(*pb.Build, pb.ReleaseOptions) (*pb.Release, error)
	BuildLogs(app, id string, index int) (int, []string, error)
	BuildLogsStream(id string) (io.Reader, error)

	ReleaseList(string, int) (pb.Releases, error)
	ReleaseDelete(string, string) error

	LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error
	ProcessRun(app string, r io.ReadWriter, command string) error

	ResourceList() (pb.Resources, error)
	ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error)
	ResourceGet(name string) (*pb.Resource, error)
	ResourceDelete(name string) error
	ResourceLink(app, name string) (*pb.Resource, error)
	ResourceUnlink(string, string) (*pb.Resource, error)

	K8sConfigPath() (string, error)
}
