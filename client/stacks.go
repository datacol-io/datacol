package client

import (
  "fmt"
  provider "github.com/dinesh/rz/cloud/google"
)

var (
  serviceKey []byte
)

func init(){
  serviceKey = []byte("serviceKey")
}

func (c *Client) CreateStack(projectId, zone, bucket string, nodes int) (*Stack, error) {
  stackName := c.StackName

  resp := provider.CreateCredential(stackName, projectId)
  if resp.Err != nil { 
    return nil, resp.Err
  }

  cred := resp.Cred
  if len(cred) == 0 {
    return nil, fmt.Errorf("invalid GCP credentials")
  }
  
  dp, err := provider.NewDeployment(cred, stackName, projectId, zone, bucket, nodes)
  if err != nil { return nil, err }

  if err := dp.Run(false); err != nil {
    if derr := dp.Delete(); derr != nil { 
      return nil, derr
    }
    return nil, err
  }

  cluster, err := dp.GetCluster()
  if err != nil { return nil, err }

  if err := provider.GenerateClusterConfig(stackName, c.configRoot(), cluster); err != nil {
    return nil, err
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

func (c *Client) DestroyStack() error {
  dp := provider.NewDeploymentByCred(c.Stack.ServiceKey)
  dp.ProjectId = c.Stack.ProjectId
  dp.Name = c.StackName

  return dp.Delete()
}
