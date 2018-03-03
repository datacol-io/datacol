package client

import (
	"io"
	"time"

	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/golang/protobuf/ptypes"
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

func (c *Client) StreamAppLogs(name string, follow bool, since time.Duration, proctype string, out io.Writer) error {
	stream, err := c.ProviderServiceClient.LogStream(ctx, &pbs.LogStreamReq{
		Name:     name,
		Since:    ptypes.DurationProto(since),
		Follow:   follow,
		Proctype: proctype,
	})
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	for {
		ret, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if _, err := out.Write(ret.Data); err != nil {
			return err
		}
	}
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
