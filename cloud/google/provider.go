package google

import (
  "context"
  "net/http"
  "log"
  "fmt"

  "github.com/dinesh/rz/client/models"

  oauth2_google "golang.org/x/oauth2/google"
  "google.golang.org/api/storage/v1"
  "google.golang.org/api/cloudbuild/v1"
  "google.golang.org/api/deploymentmanager/v2"
  "google.golang.org/api/container/v1"
  csm "google.golang.org/api/cloudresourcemanager/v1"
)

var (
  appBucket = "App"
)

var _jwtClient *http.Client

type GCPCloud struct {
  Project        string
  DeploymentName string
  BucketName     string
  Zone           string
  ServiceKey     []byte
}

func (g *GCPCloud) AppCreate(app *models.App) error {
 
  return nil
}

func (g *GCPCloud) AppGet(name string) (*models.App, error) {
  return nil, nil 
}

func (g *GCPCloud) AppDelete(name string) error {
  return nil
}

func (g *GCPCloud) storage() *storage.Service {
  svc, err := storage.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("storage client %s", err))
  }

  return svc
}

func (g *GCPCloud) cloudbuilder() *cloudbuild.Service {
  svc, err := cloudbuild.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("cloudbuilder client %s", err))
  }

  return svc
}

func (g *GCPCloud) csmanager() *csm.Service {
  svc, err := csm.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("cloudresourcemanager client %s", err))
  }

  return svc
}

func (g *GCPCloud) deploymentmanager() *deploymentmanager.Service {
  svc, err := deploymentmanager.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("deploymentmanager client %s", err))
  }

  return svc
}

func (g *GCPCloud) container() *container.Service {
  svc, err := container.New(jwtClient(g.ServiceKey))
  if err != nil {
    log.Fatal(fmt.Errorf("container client %s", err))
  }

  return svc
}

func (g *GCPCloud) projectNumber() (int64, error) {
  service := g.csmanager()
  op, err := service.Projects.Get(g.Project).Do()

  if err != nil {
    return 0, fmt.Errorf("fetching %s: %s",g.Project,err)
  }

  return op.ProjectNumber, nil
}

func (g *GCPCloud) getCluster(name string) (*container.Cluster, error) {
  service := g.container()
  return service.Projects.Zones.Clusters.Get(g.Project, g.Zone, name).Do()
}


func jwtClient(sva []byte) *http.Client {
  if _jwtClient != nil {
    return _jwtClient
  }

  jwtConfig, err := oauth2_google.JWTConfigFromJSON(sva, csm.CloudPlatformScope)
  if err != nil {
    log.Fatal(fmt.Errorf("JWT client %s", err))
  }

  _jwtClient = jwtConfig.Client(context.Background())
  return _jwtClient
}

