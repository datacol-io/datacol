package client

import (
	"fmt"
	log "github.com/Sirupsen/logrus"

	"os"
	"path/filepath"
	"strings"

	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/api/models"

	"github.com/dinesh/datacol/cloud"
	"github.com/dinesh/datacol/cmd/stdcli"
	"google.golang.org/grpc"
)

func init() {
	root := models.ConfigPath
	if _, err := os.Stat(root); err != nil {
		if !os.IsNotExist(err) {
			stdcli.Error(err)
			return
		} else {
			if err := os.MkdirAll(root, 0700); err != nil {
				stdcli.Error(err)
				return
			}
		}
	}
}

type Client struct {
	Version   string
	StackName string
	ProjectId string
}

func (c *Client) Stack() (pb.StackServiceClient, func() error) {
	conn, err := grpc.Dial(":8080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(fmt.Errorf("did not connect: %v", err))
	}
	
	return pb.NewStackServiceClient(conn), conn.Close
}

func (c *Client) configRoot() string {
	return filepath.Join(models.ConfigPath, c.StackName)
}

func (c *Client) SetStack(name string) {
	parts := strings.Split(name, "@")

	c.StackName = parts[0]
	if len(parts) > 1 {
		c.ProjectId = parts[1]
	} else {
		c.ProjectId = os.Getenv("PROJECT_ID")
	}

	if len(c.ProjectId) == 0 {
		log.Fatal(fmt.Errorf("GCP project-id not found. Please set `PROJECT_ID` environment variable."))
	}
}

func (c *Client) Provider() cloud.Provider {
	if len(c.StackName) == 0 || len(c.ProjectId) == 0 {
		log.Fatal(stdcli.Stack404)
	}

	return cloud.Getgcp(c.StackName, c.ProjectId)
}
