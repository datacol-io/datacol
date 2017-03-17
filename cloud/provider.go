package cloud

import (
  "github.com/dinesh/rz/cloud/google"
  "github.com/dinesh/rz/client/models"

)

type Provider interface {
  Initialize(string, int) error
  Teardown() error

  AppCreate(app *models.App) error
  AppGet(name string) (*models.App, error)
  AppDelete(name string) error

  BuildImport(key string, source []byte) error
  BuildCreate(app, source string, opts *models.BuildOptions) error

  // BuildDelete(app, id string) (*client.Build, error)
  // BuildGet(app, id string) (*client.Build, error)
  // BuildLogs(app, id string, w io.Writer) error
  // BuildList(app string, limit int64) (client.Builds, error)
  // BuildRelease(*client.Build) (*client.Release, error)
  // BuildSave(*client.Build) error
}


func Getgcp(name, project, zone, bucket string, serviceKey []byte) Provider {
  return &google.GCPCloud{
    Project:  project,
    Zone:     zone,
    BucketName:   bucket,
    DeploymentName: name,
    ServiceKey: serviceKey,
  }
}