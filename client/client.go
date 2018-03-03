package client

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/io"
	term "github.com/appscode/go/term"
	pb "github.com/datacol-io/datacol/api/controller"
	"github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	apiHttpPort = 8080
	apiRpcPort  = 10000
	apiTimeout  = 10
)

func init() {
	root := models.ConfigPath
	if err := io.EnsureDirectory(root); err != nil {
		stdcli.ExitOnError(err)
	}
}

type Client struct {
	Version string
	pb.ProviderServiceClient
	stdcli.Auth
}

func (c *Client) IsGCP() bool {
	return c.Provider == "gcp"
}

func (c *Client) IsAWS() bool {
	return c.Provider == "aws"
}

func (c *Client) IsLocal() bool {
	return c.Provider == "local"
}

func NewClient(version string) (*Client, func() error) {
	auth, _ := stdcli.GetAuthOrDie()

	psc, close := GrpcClient(auth.ApiServer, auth.ApiKey)
	conn := &Client{
		Version:               version,
		ProviderServiceClient: psc,
		Auth: *auth,
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
		grpc.WithBlock(),
		grpc.WithTimeout(time.Second*apiTimeout),
		grpc.WithPerRPCCredentials(&loginCreds{ApiKey: password}),
	)
	if err != nil {
		if err == grpc.ErrClientConnTimeout || err == context.DeadlineExceeded {
			term.Errorln("Couldn't connect to the Controller API. Did you initialize the stack using `datacol init`")
		}

		stdcli.ExitOnError(err)
	}

	return pb.NewProviderServiceClient(conn), conn.Close
}

func (c *Client) SetStack(auth *stdcli.Auth) {
	c.Auth = *auth

	if c.IsGCP() {
		// for GCP only
		if len(c.Project) == 0 {
			c.Project = stdcli.ReadSetting(c.Name, "project")
		}

		if len(auth.Project) == 0 && len(auth.Region) == 0 {
			log.Fatal(fmt.Errorf("GCP project-id not found. Please set `PROJECT_ID` environment variable."))
		}
	}
}
