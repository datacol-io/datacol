package client

import (
	"errors"
	pb "github.com/dinesh/datacol/api/models"
	provider "github.com/dinesh/datacol/cloud/google"
	"github.com/dinesh/datacol/cmd/stdcli"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrNotFound = errors.New("key not found")
	credNotFound = errors.New("Invalid credentials")
)

func (c *Client) CreateStack(project, zone, bucket string, optout bool) (*stack, error) {
	stackName := c.StackName

	resp := provider.CreateCredential(stackName, project, optout)
	if resp.Err != nil {
		return nil, resp.Err
	}

	cred := resp.Cred
	if len(cred) == 0 {
		return nil, credNotFound
	}

	if err := os.MkdirAll(c.configRoot(), 0777); err != nil {
		return nil, err
	}

	if err := saveCredential(stackName, cred); err != nil {
		return nil, err
	}

	st := &pb.Stack{
		Name:       stackName,
		Zone:       zone,
		Bucket:     bucket,
		ServiceKey: cred,
		ProjectId:  resp.ProjectId,
		ProjectNumber: resp.PNumber,
	}

	time.Sleep(2 * time.Second)

	// if err := c.Provider().StackSave(st); err != nil {
	// 	return nil, err
	// }

	return &stack{Name: st.Name, ProjectId: st.ProjectId}, nil
}

func (c *Client) DeployStack(st *stack, clusterName, machineType string, nodes int, preem bool) error {
	return nil
	// return c.Provider().Initialize(clusterName, machineType, nodes, preem)
}

func (c *Client) DestroyStack() error {
	// if err := c.Provider().Teardown(); err != nil {
	// 	return err
	// }

	return nil
}

func (c *Client) SetFromEnv() {
	c.SetStack(stdcli.GetStack())
}

func saveCredential(name string, data []byte) error {
	path := filepath.Join(pb.ConfigPath, name, pb.SvaFilename)
	return ioutil.WriteFile(path, data, 0777)
}
