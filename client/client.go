package client

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"

	"github.com/appscode/go/io"
	pb "github.com/dinesh/datacol/api/controller"
	"github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cmd/stdcli"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	apiHttpPort = 8080
	apiRpcPort  = 10000
)

func init() {
	root := models.ConfigPath
	if err := io.EnsureDirectory(root); err != nil {
		stdcli.Error(err)
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

	v, err := io.ReadFile(filepath.Join(models.ConfigPath, name, "api_host"))
	if err != nil {
		log.Fatal(err)
	}
	host := strings.TrimSpace(v)

	p, err := io.ReadFile(filepath.Join(models.ConfigPath, name, "api_key"))
	if err != nil {
		log.Fatal(err)
	}
	password := strings.TrimSpace(p)

	psc, close := GrpcClient(host, password)
	conn := &Client{
		Version:               version,
		ProviderServiceClient: psc,
	}
	conn.SetStack(name)

	return conn, close
}

type loginCreds struct {
	ApiKey string
}

func (lc *loginCreds) GetRequestMetadata(c context.Context, args ...string) (map[string]string, error) {
	return map[string]string{
		"api_key": lc.ApiKey,
	}, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return false
}

func GrpcClient(host, password string) (pb.ProviderServiceClient, func() error) {
	address := fmt.Sprintf("%s:%d", host, apiRpcPort)
	log.Debugf("grpc dialing at %s", address)

	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(&loginCreds{ApiKey: password}))
	if err != nil {
		log.Fatal(fmt.Errorf("did not connect: %v", err))
	}

	return pb.NewProviderServiceClient(conn), conn.Close
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
