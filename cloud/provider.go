package cloud

import (
	"io"

	"github.com/dinesh/datacol/client/models"
	"github.com/dinesh/datacol/cloud/google"
)

type Provider interface {
	Initialize(string, string, int, bool) error
	StackSave(*models.Stack) error
	Teardown() error

	AppCreate(string) (*models.App, error)
	AppGet(string) (*models.App, error)
	AppDelete(string) error
	AppRestart(string) error
	AppList() (models.Apps, error)

	BuildImport(key string, source []byte) error
	BuildCreate(app, source string, opts *models.BuildOptions) error

	EnvironmentGet(app string) (models.Environment, error)
	EnvironmentSet(app string, body io.Reader) error

	LogStream(app string, w io.Writer, opts models.LogStreamOptions) error

	BuildList(app string, limit int) (models.Builds, error)
	BuildGet(app, id string) (*models.Build, error)
	// BuildSave(*client.Build) error
	BuildDelete(app, id string) error
	// BuildLogs(app, id string, w io.Writer) error
	BuildRelease(*models.Build) (*models.Release, error)

	ReleaseList(string, int) (models.Releases, error)
	ReleaseDelete(string, string) (error)

	ResourceCreate(name, kind string, params map[string]string) (*models.Resource, error)
	ResourceDelete(name string) error
	ResourceGet(name string) (*models.Resource, error)
	ResourceLink(app, name string) (*models.Resource, error)
	ResourceList() (models.Resources, error)
	ResourceUnlink(string, string) (*models.Resource, error)
	// ResourceUpdate(name string, params map[string]string) (*structs.Resource, error)

	GetRunningPods(string) (string, error)
}

func Getgcp(name, project string) Provider {
	return &google.GCPCloud{
		Project:        project,
		DeploymentName: name,
	}
}
