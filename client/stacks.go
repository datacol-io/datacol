package client

import (
  "fmt"
  "errors"
  provider "github.com/dinesh/rz/cloud/google"
)

var (
  credNotFound = errors.New("Invalid credentials")
)

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

func (c *Client) DeployStack(st *Stack, clusterName string, nodes int) error {
  if len(st.ServiceKey) == 0 {
    return credNotFound
  }
  provider := c.Provider()

  if err := provider.Initialize(clusterName, nodes); err != nil {
    fmt.Printf("failed: %v\n", err)
    if derr := provider.Teardown(); derr != nil { 
      return derr
    }
    return err
  }
  return nil
}

func (c *Client) DestroyStack() error {
  return c.Provider().Teardown()
}
