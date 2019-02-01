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
	pb "github.com/datacol-io/datacol/api/controller"
	"github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

const (
	apiHttpPort = 8080
	apiRpcPort  = 10000
	apiTimeout  = 20
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

func NewClient(c *cli.Context) (*Client, func() error) {
	auth, _ := stdcli.GetAuthOrDie(c)

	psc, close := GrpcClient(auth.ApiServer, auth.ApiKey)
	conn := &Client{
		Version: c.App.Version,
		Auth:    *auth,
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
			stdcli.ExitOnErrorf("Couldn't connect to %s", address)
		}

		stdcli.ExitOnError(err)
	}

	return pb.NewProviderServiceClient(conn), conn.Close
}

func (c *Client) Stream(path string, headers map[string]string, in io.ReadCloser, out io.Writer) error {
	ws, err := c.StreamClient(path, headers)
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

func (c *Client) StreamClient(path string, headers map[string]string) (*websocket.Conn, error) {
	return wsConn(path, c.ApiServer, c.Version, c.ApiKey, headers)
}

func wsConn(path, server, version, apiKey string, headers map[string]string) (*websocket.Conn, error) {
	origin := fmt.Sprintf("https://%s:%d", server, apiHttpPort)
	endpoint := fmt.Sprintf("ws://%s:%d%s", server, apiHttpPort, path)

	config, err := websocket.NewConfig(endpoint, origin)

	if err != nil {
		return nil, err
	}

	config.TlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	config.Header.Set("Version", version)

	userpass := fmt.Sprintf(":%s", apiKey)
	userpass_encoded := base64.StdEncoding.EncodeToString([]byte(userpass))

	config.Header.Add("Authorization", fmt.Sprintf("Basic %s", userpass_encoded))

	for k, v := range headers {
		config.Header.Add(k, v)
	}

	config.TlsConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	return websocket.DialConfig(config)
}

func copyAsync(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	io.Copy(dst, src)
}

func iocopy(dst io.Writer, src io.Reader, c chan error) {
	_, err := io.Copy(dst, src)
	c <- err
}
