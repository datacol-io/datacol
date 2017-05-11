package client

import (
	"fmt"
	log "github.com/Sirupsen/logrus"

	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/dinesh/datacol/api/controller"
	"github.com/dinesh/datacol/api/models"

	"github.com/dinesh/datacol/cloud"
	"github.com/dinesh/datacol/cmd/stdcli"
	"google.golang.org/grpc"
)

const (
	apiHttpPort = 8080
	apiRpcPort  = 10000
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

	pb.ProviderServiceClient
}

func (c *Client) SetFromEnv() {
	c.SetStack(stdcli.GetAppStack())
}

func NewClient(version string) (*Client, func() error) {
	name := stdcli.GetAppStack()
	v, err := ioutil.ReadFile(filepath.Join(models.ConfigPath, name, "api_host"))
	if err != nil {
		log.Fatal(err)
	}
	host := strings.TrimSpace(string(v))

	psc, close := GrpcClient(host)
	conn := &Client{
		Version:               version,
		ProviderServiceClient: psc,
	}
	conn.SetStack(name)

	return conn, close
}

func GrpcClient(host string) (pb.ProviderServiceClient, func() error) {
	address := fmt.Sprintf("%s:%d", host, apiRpcPort)
	log.Debugf("grpc dialing at %s", address)

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatal(fmt.Errorf("did not connect: %v", err))
	}

	return pb.NewProviderServiceClient(conn), conn.Close
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
		c.ProjectId = stdcli.ReadSetting(c.StackName, "project")
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
