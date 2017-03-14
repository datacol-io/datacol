package client

import (
  "errors"
  provider "github.com/dinesh/rz/cloud/google"
)

var (
  credNotFound = errors.New("Invalid credentials")
)

func init(){
}

func (c *Client) CreateStack(projectId, zone, bucket string) (*Stack, error) {
  stackName := c.StackName

  resp := provider.CreateCredential(stackName, projectId)
  if resp.Err != nil {
    return nil, resp.Err
  }

  cred := resp.Cred
  if len(cred) == 0 {
    return nil, credNotFound
  }

  st := &Stack{
    Name:       stackName,
    ProjectId:  projectId,
    Zone:       zone,
    Bucket:     bucket,
    ServiceKey: cred,
  }

  if err := st.Persist(true); err != nil { 
    return nil, err
  }
  return st, nil
}

func (c *Client) DeployStack(st *Stack, nodes int) error {  
  dp, err := provider.NewDeployment(st.ServiceKey, st.Name, st.ProjectId, st.Zone, st.Bucket, nodes)
  if err != nil { return err }

  if err := dp.Run(false); err != nil {
    if derr := dp.Delete(); derr != nil { 
      return derr
    }
    return err
  }

  cluster, err := dp.GetCluster()
  if err != nil { return err }

  return provider.GenerateClusterConfig(st.Name, c.configRoot(), cluster)
}

func (c *Client) DestroyStack() error {
  dp := provider.NewDeploymentByCred(c.Stack.ServiceKey)
  dp.ProjectId = c.Stack.ProjectId
  dp.Name = c.StackName

  return dp.Delete()
}
