package cloud

import (
	"io"

	"github.com/dinesh/datacol/client/models"
	"github.com/dinesh/datacol/cloud/google"
)

type Provider interface {
	Initialize(string, string, int, bool) error
	Teardown() error

	AppCreate(app *models.App) error
	AppGet(string) (*models.App, error)
	AppDelete(string) error
	AppRestart(string) error

	BuildImport(key string, source []byte) error
	BuildCreate(app, source string, opts *models.BuildOptions) error

	EnvironmentGet(app string) (models.Environment, error)
	EnvironmentSet(app string, body io.Reader) error

	LogStream(app string, w io.Writer, opts models.LogStreamOptions) error

	// BuildDelete(app, id string) (*client.Build, error)
	// BuildGet(app, id string) (*client.Build, error)
	// BuildLogs(app, id string, w io.Writer) error
	// BuildList(app string, limit int64) (client.Builds, error)
	// BuildRelease(*client.Build) (*client.Release, error)
	// BuildSave(*client.Build) error

	ResourceCreate(name, kind string, params map[string]string) (*models.Resource, error)
	ResourceDelete(name string) error
	// ResourceGet(name string) (*structs.Resource, error)
	ResourceLink(app string, resource *models.Resource) (*models.Resource, error)
	ResourceList() (models.Resources, error)
	ResourceUnlink(string, *models.Resource) (*models.Resource, error)
	// ResourceUpdate(name string, params map[string]string) (*structs.Resource, error)

	GetRunningPods(string) (string, error)
}

func Getgcp(name, project string, pnumber int64, zone, bucket string, serviceKey []byte) Provider {
	return &google.GCPCloud{
		Project:        project,
		ProjectNumber:  pnumber,
		Zone:           zone,
		BucketName:     bucket,
		DeploymentName: name,
		ServiceKey:     serviceKey,
	}
}
