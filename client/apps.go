package client

import (
	"io"
	"time"
  "golang.org/x/net/context"
  pb "github.com/dinesh/datacol/api/models"
  pbs "github.com/dinesh/datacol/api/controller"
)

var ctx = context.TODO()

func (c *Client) GetApps() (pb.Apps, error) {
	ret, err := c.ProviderServiceClient.AppList(ctx, &pbs.ListRequest{})
  return ret.Apps, err
}

func (c *Client) GetApp(name string) (*pb.App, error) {
	return c.ProviderServiceClient.AppGet(ctx, &pbs.AppRequest{Name: name})
}

func (c *Client) CreateApp(name string) (*pb.App, error) {
	return c.ProviderServiceClient.AppCreate(ctx, &pbs.AppRequest{Name: name})
}

func (c *Client) DeleteApp(name string) error {
	_, err := c.ProviderServiceClient.AppDelete(ctx, &pbs.AppRequest{Name: name})
  return err
}

func (c *Client) StreamAppLogs(name string, follow bool, since time.Duration, out io.Writer) error {
	opts := pb.LogStreamOptions{Since: since, Follow: follow}
	return c.Provider().LogStream(name, out, opts)
}

func (c *Client) GetEnvironment(name string) (pb.Environment, error) {
  ret, err := c.ProviderServiceClient.EnvironmentGet(ctx, &pbs.AppRequest{Name: name})
  if err != nil { return nil, err }
  return ret.Data, nil
}

func (c *Client) SetEnvironment(name string, data string) error {
  _, err := c.ProviderServiceClient.EnvironmentSet(ctx, &pbs.EnvSetRequest{Name: name, Data: data})
  return err
}
