package client

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"golang.org/x/net/context"
)

var (
	ctx = context.TODO()
)

func (c *Client) GetApps() (pb.Apps, error) {
	ret, err := c.ProviderServiceClient.AppList(ctx, &pbs.ListRequest{})
	return ret.Apps, err
}

func (c *Client) GetApp(name string) (*pb.App, error) {
	return c.ProviderServiceClient.AppGet(ctx, &pbs.AppRequest{Name: name})
}

func (c *Client) CreateApp(name, repo string) (*pb.App, error) {
	return c.ProviderServiceClient.AppCreate(ctx, &pbs.AppRequest{
		Name:    name,
		RepoUrl: repo,
	})
}

func (c *Client) DeleteApp(name string) error {
	_, err := c.ProviderServiceClient.AppDelete(ctx, &pbs.AppRequest{Name: name})
	return err
}

func (c *Client) RestartApp(name string) error {
	_, err := c.ProviderServiceClient.AppRestart(ctx, &pbs.AppRequest{Name: name})
	return err
}

func (c *Client) StreamAppLogs(name string, follow bool, since time.Duration, process string, lines int, out io.Writer) error {
	in, out := os.Stdin, os.Stdout
	return c.Stream("/ws/v1/logs", map[string]string{
		"app":     name,
		"since":   since.String(),
		"follow":  strconv.FormatBool(follow),
		"Process": process,
		"lines":   strconv.Itoa(lines),
	}, in, out)
}

func (c *Client) GetEnvironment(name string) (pb.Environment, error) {
	ret, err := c.ProviderServiceClient.EnvironmentGet(ctx, &pbs.AppRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return ret.Data, nil
}

func (c *Client) SetEnvironment(name string, data string) error {
	_, err := c.ProviderServiceClient.EnvironmentSet(ctx, &pbs.EnvSetRequest{Name: name, Data: data})
	return err
}

func (c *Client) ProxyRemote(host string, port int, conn io.ReadWriteCloser) error {
	return c.Stream("/ws/v1/proxy", map[string]string{
		"remotehost": host,
		"remoteport": fmt.Sprintf("%d", port),
	}, conn, conn)
}
