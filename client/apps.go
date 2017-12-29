package client

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
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

func (c *Client) StreamAppLogs(name string, follow bool, since time.Duration, out io.Writer) error {
	stream, err := c.ProviderServiceClient.LogStream(ctx, &pbs.LogStreamReq{
		Name:   name,
		Since:  ptypes.DurationProto(since),
		Follow: follow,
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

func (c *Client) ListProcess(name string) ([]*pb.Process, error) {
	resp, err := c.ProviderServiceClient.ProcessList(ctx, &pbs.AppRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func (c *Client) RunProcess(name string, args []string) error {
	newctx := metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"app":     name,
		"command": strings.Join(args, " "),
	}))

	stream, err := c.ProviderServiceClient.ProcessRun(newctx)
	if err != nil {
		return err //FIXME: not able to make it work for now. Not sure why.
	}

	defer stream.CloseSend()

	wg := sync.WaitGroup{}
	wg.Add(1)

	r, w := os.Stdin, os.Stdout

	go func(out io.Writer) {
		defer wg.Done()

		for {
			ret, rerr := stream.Recv()
			if rerr == io.EOF {
				break
			}

			if rerr != nil {
				err = rerr
				break
			}

			if _, err = out.Write(ret.Data); err != nil {
				break
			}
		}
	}(w)

	go func(r io.Reader) {
		defer wg.Done()
		buf := make([]byte, 1024*1024)

		for {
			n, serr := r.Read(buf)
			if serr == io.EOF {
				break
			}
			if serr != nil {
				err = serr
				break
			}

			if serr := stream.Send(&pbs.StreamMsg{Data: buf[:n]}); serr != nil {
				err = serr
				break
			}
		}
	}(r)

	wg.Wait()

	return err
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
