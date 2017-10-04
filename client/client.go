package client

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
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
		stdcli.ExitOnError(err)
	}
}

type Client struct {
	Version   string
	StackName string
	ProjectId string

	pb.ProviderServiceClient
}

func (c *Client) IsGCP() bool {
	return len(c.ProjectId) > 0
}

func (c *Client) IsAWS() bool {
	return !c.IsGCP()
}

func (c *Client) Provider() string {
	if c.IsAWS() {
		return "AWS"
	}
	if c.IsGCP() {
		return "GCP"
	}

	return "Unknown"
}

func NewClient(version string) (*Client, func() error) {
	auth, _ := stdcli.GetAuthOrDie()

	psc, close := GrpcClient(auth.ApiServer, auth.ApiKey)
	conn := &Client{
		Version:               version,
		ProviderServiceClient: psc,
	}

	conn.SetStack(auth)
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

	conn, err := grpc.Dial(address,
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(&loginCreds{ApiKey: password}))
	if err != nil {
		log.Fatal(fmt.Errorf("did not connect: %v", err))
	}

	return pb.NewProviderServiceClient(conn), conn.Close
}

func (c *Client) SetStack(auth *stdcli.Auth) {
	c.StackName = auth.Name
	c.ProjectId = auth.Project

	if len(auth.Region) == 0 {
		// for GCP only
		if len(c.ProjectId) == 0 {
			c.ProjectId = os.Getenv("PROJECT_ID")
		}

		if len(c.ProjectId) == 0 {
			c.ProjectId = stdcli.ReadSetting(c.StackName, "project")
		}

		if len(c.ProjectId) == 0 && len(auth.Region) == 0 {
			log.Fatal(fmt.Errorf("GCP project-id not found. Please set `PROJECT_ID` environment variable."))
		}
	}
}
