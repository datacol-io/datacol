package client

import (
  "os"
  "errors"
  provider "github.com/dinesh/datacol/cloud/google"
)

var (
  credNotFound = errors.New("Invalid credentials")
)

func (c *Client) CreateStack(project, zone, bucket string) (*Stack, error) {
  stackName := c.StackName

  resp := provider.CreateCredential(stackName, project)
  if resp.Err != nil {
    return nil, resp.Err
  }

  cred := resp.Cred
  if len(cred) == 0 {
    return nil, credNotFound
  }

  err := os.MkdirAll(c.configRoot(), 0777); if err != nil {
    return nil, err
  }

  st := &Stack{
    Name:       stackName,
    ProjectId:  resp.ProjectId,
    Zone:       zone,
    Bucket:     bucket,
    ServiceKey: cred,
  }

  if err := st.Persist(true); err != nil { 
    return nil, err
  }
  return st, nil
}

func (c *Client) DeployStack(st *Stack, clusterName, machineType string, nodes int, preem bool) error {
  if len(st.ServiceKey) == 0 {
    return credNotFound
  }

  return c.Provider().Initialize(clusterName, machineType, nodes, preem)
}

func (c *Client) DestroyStack() error {
  if err := c.Provider().Teardown(); err != nil {
    return err
  }
  return c.purgeStack()
}

func (c *Client) purgeStack() error {
  name := c.Stack.Name
  apps, err := c.GetApps()
  if err != nil { return err }
  
  for _, app := range apps {
    builds, err := c.GetBuilds(app.Name)
    if err != nil { return err }

    for _, b := range builds {
      if err := c.DeleteBuild(b.Id); err != nil {
        return err
      }
    }

    rs, err := c.GetReleases(app.Name)
    if err != nil { return err }

    for _, r := range rs {
      if err := c.DeleteRelease(r.Id); err != nil {
        return err
      }
    }
    c.DeleteApp(app.Name)
  }

  sbx, _ := DB.New(stackBxName)
  if err := sbx.Delete([]byte(name)); err != nil {
    return err
  }

  return os.RemoveAll(c.configRoot())
}
