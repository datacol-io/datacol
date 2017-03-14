package client

import (
  "time"
  "fmt"
  "path/filepath"
  "k8s.io/client-go/pkg/util/intstr"

  provider "github.com/dinesh/rz/cloud/google"
)

type Release struct {
  Id      string  `json: "id"`
  App     string  `json: "app"`
  BuildId string  `json: "buildId"`
  Status  string  `json: "status"`
  CreatedAt time.Time `json: "created_at"`
}

func (c *Client) NewRelease(b *Build) *Release {
  r := &Release {
    Id:         generateId("R", 5),
    App:        b.App, 
    BuildId:    b.Id,
    CreatedAt:  time.Now(),
  }

  return r
}

func (c *Client) DeployRelease(r *Release, port int, image, env string) error {
  if image == "" {
    image = fmt.Sprintf("gcr.io/%v/%v", c.Stack.ProjectId, r.App)
  }
  if env == "" {
    env = c.Stack.Name
  }

  cfgpath := filepath.Join(c.configRoot(), "kubeconfig")
  // cfgpath := filepath.Join(os.Getenv("HOME") + "/.kube/config")

  token, err := c.PrdToken()
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

  fmt.Printf("Deployed at port: %d", resp.NodePort)
  return nil
}

