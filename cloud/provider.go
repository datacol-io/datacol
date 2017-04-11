package cloud

import (
  "io"

  "github.com/dinesh/datacol/cloud/google"
  "github.com/dinesh/datacol/client/models"
)

type Provider interface {
  Initialize(string, string, int, bool) error
  Teardown() error

  AppCreate(app *models.App) error
  AppGet(name string) (*models.App, error)
  AppDelete(name string) error

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