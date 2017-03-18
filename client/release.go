package client

import (
  "time"
  "fmt"
  "path/filepath"
  "encoding/json"
  "k8s.io/client-go/pkg/util/intstr"
  "github.com/dinesh/rz/client/models"

  provider "github.com/dinesh/rz/cloud/google"
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

func (c *Client) DeployRelease(r *models.Release, port int, image, env string) error {
  if image == "" {
    image = fmt.Sprintf("gcr.io/%v/%v:%v", c.Stack.ProjectId, r.App, r.BuildId)
  }

  if env == "" {
    env = c.Stack.Name
  }

  cfgpath := filepath.Join(c.configRoot(), "kubeconfig")

  token, err := c.Provider().BearerToken()
  if err != nil { return err }

  deployer, err := provider.NewDeployer(cfgpath, token)
  if err != nil { 
    return err
  }

  req := &provider.DeployRequest{
    ServiceID:     r.App,
    Image:         image,
    Replicas:      1,
    Environment:   env,
    Zone:          c.Stack.Zone,
    ContainerPort: intstr.FromInt(port),
  }

  resp, err := deployer.Run(req)
  if err != nil {
    return err
  }

  r.Status = "success"
  if err := Persist(r_bucket, r.Id, r); err != nil {
    return err
  }

  fmt.Printf("Deployed at port: %d", resp.NodePort)
  return nil
}

