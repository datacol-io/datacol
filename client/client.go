package client

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	io_ext "github.com/appscode/go/io"
	term "github.com/appscode/go/term"
	pb "github.com/datacol-io/datacol/api/controller"
	"github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

const (
	apiHttpPort = 8080
	apiRpcPort  = 10000
	apiTimeout  = 10
)

func init() {
	root := models.ConfigPath
	if err := io_ext.EnsureDirectory(root); err != nil {
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
		Version: version,

		Auth: *auth,
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

func (c *Client) Stream(path string, headers map[string]string, in io.Reader, out io.Writer) error {
	origin := fmt.Sprintf("https://%s:%d", c.ApiServer, apiHttpPort)
	endpoint := fmt.Sprintf("ws://%s:%d%s", c.ApiServer, apiHttpPort, path)

	config, err := websocket.NewConfig(endpoint, origin)

	if err != nil {
		return err
	}

	config.TlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	config.Header.Set("Version", c.Version)

	userpass := fmt.Sprintf(":%s", c.ApiKey)
	userpass_encoded := base64.StdEncoding.EncodeToString([]byte(userpass))

	config.Header.Add("Authorization", fmt.Sprintf("Basic %s", userpass_encoded))

	for k, v := range headers {
		config.Header.Add(k, v)
	}

	config.TlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	ws, err := websocket.DialConfig(config)
	if err != nil {
		return err
	}
	defer ws.Close()

	var wg sync.WaitGroup

	if in != nil {
		go io.Copy(ws, in)
	}

	if out != nil {
		wg.Add(1)
		go copyAsync(out, ws, &wg)
	}

	wg.Wait()

	return nil
}

func copyAsync(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	io.Copy(dst, src)
}
