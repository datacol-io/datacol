package cloud

import (
	"io"

	pb "github.com/datacol-io/datacol/api/models"
)

type Provider interface {
	AppCreate(name string, req *pb.AppCreateOptions) (*pb.App, error)
	AppGet(string) (*pb.App, error)
	AppDelete(string) error
	AppRestart(string) error
	AppList() (pb.Apps, error)
	AppUpdateDomain(string, string) error

	CertificateCreate(app, domain, cert, key string) error
	CertificateDelete(app, domain string) error

	EnvironmentGet(app string) (pb.Environment, error)
	EnvironmentSet(app string, body io.Reader) error

	BuildCreate(app string, req *pb.CreateBuildOptions) (*pb.Build, error)
	BuildUpload(app, filename string) error
	BuildImport(app string, r io.Reader, w io.WriteCloser) error
	BuildList(app string, limit int64) (pb.Builds, error)
	BuildGet(app, id string) (*pb.Build, error)
	BuildDelete(app, id string) error
	BuildRelease(*pb.Build, pb.ReleaseOptions) (*pb.Release, error)
	BuildLogs(app, id string, index int) (int, []string, error)
	BuildLogsStream(id string) (io.Reader, error)

	ReleaseList(string, int64) (pb.Releases, error)
	ReleaseDelete(string, string) error

	LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error

	ProcessList(app string) ([]*pb.Process, error)
	ProcessRun(app string, r io.ReadWriter, command []string) error
	ProcessSave(app string, formation map[string]int32) error
	ProcessLimits(app, resource string, limits map[string]string) error

	ResourceList() (pb.Resources, error)
	ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error)
	ResourceGet(name string) (*pb.Resource, error)
	ResourceDelete(name string) error
	ResourceLink(app, name string) (*pb.Resource, error)
	ResourceUnlink(string, string) (*pb.Resource, error)

	K8sConfigPath() (string, error)
	DockerCredsGet() (*pb.DockerCred, error)
}
