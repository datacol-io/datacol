package client

import (
  "time"
  "fmt"
  "encoding/json"
  "k8s.io/client-go/pkg/util/intstr"
  "github.com/dinesh/datacol/client/models"
  "github.com/dinesh/datacol/cloud/google"
  log "github.com/Sirupsen/logrus"
)

var (
  r_bucket = []byte("releases")
)

func (c *Client) NewRelease(b *models.Build) *models.Release {
  r := &models.Release {
    Id:         generateId("R", 5),
    App:        b.App, 
    BuildId:    b.Id,
    CreatedAt:  time.Now(),
  }

  return r
}

func (c *Client) GetReleases(app string) (models.Releases, error) {
  rbx, _ := DB.New(r_bucket)
  items, err := rbx.Items()
  if err != nil { return nil, err }

  var rs models.Releases

  for _, item := range items {
    var r models.Release
    err := json.Unmarshal(item.Value, &r)
    if err != nil { return nil, err }

    if r.App == app {
      rs = append(rs, &r)
    }
  }

  return rs, nil
}

func (c *Client) DeleteRelease(Id string) error {
  rbx, _ := DB.New(r_bucket)
  return rbx.Delete([]byte(Id))
}

func (c *Client) DeployRelease(r *models.Release, port int, image, env string, wait bool) error {
  if image == "" {
    image = fmt.Sprintf("gcr.io/%v/%v:%v", c.Stack.ProjectId, r.App, r.BuildId)
  }

  if env == "" {
    env = c.Stack.Name
  }

  provider := c.Provider()
  envVars, err := provider.EnvironmentGet(r.App)
  if err != nil {
    return err
  }

  deployer, err := google.NewDeployer(env)
  if err != nil { 
    return err
  }

  req := &google.DeployRequest{
    ServiceID:     r.App,
    Image:         image,
    Replicas:      1,
    Environment:   env,
    Zone:          c.Stack.Zone,
    ContainerPort: intstr.FromInt(port),
    EnvVars:       envVars,
  }

  if _, err := deployer.Run(req); err != nil {
    return err
  }

  r.Status = "success"
  if err := Persist(r_bucket, r.Id, r); err != nil {
    return err
  }

  app, err := c.GetApp(r.App)
  if err != nil { return err }

  app.Status = "Running"
  if err = Persist(a_bucket, app.Name, &app); err != nil {
    return err
  }

  if wait {
    waitTill := time.Now().Add(time.Duration(1) * time.Minute)
    fmt.Printf("waiting for ip")

    for {
      time.Sleep(1 * time.Second)
      fmt.Print(".")
      if err = c.SyncApp(app, wait); err != nil {
        return err
      }

      if len(app.HostPort) > 0 {
        break
      }

      if time.Now().After(waitTill) {
        log.Warn("timed out. Skipping.")
        break
      }
    }
  }

  return nil
}

