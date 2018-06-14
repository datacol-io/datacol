package client

import (
	"fmt"
	"io"
	"math"

	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"google.golang.org/grpc/metadata"
)

const chunkSize = 1024 * 1024 * 1

func (c *Client) CreateBuild(app *pb.App, data []byte, procfile []byte) (*pb.Build, error) {
	b, err := c.ProviderServiceClient.BuildCreate(ctx, &pbs.CreateBuildRequest{
		App:      app.Name,
		Procfile: procfile,
	})

	if err != nil {
		return nil, err
	}

	md := metadata.Pairs("app", app.Name, "id", b.Id)
	newctx := metadata.NewOutgoingContext(ctx, md)

	stream, err := c.ProviderServiceClient.BuildImport(newctx)
	defer stream.CloseSend()

	if err != nil {
		return nil, err
	}

	numChunks := int(math.Ceil(float64(len(data)) / float64(chunkSize)))

	fmt.Print("Uploading source")

	for i := 0; i < numChunks; i++ {
		maxEnd := i*chunkSize + intMin(chunkSize, len(data[i*chunkSize:]))

		chunk := data[i*chunkSize : maxEnd]
		if len(chunk) == 0 && err == io.EOF {
			break
		}

		fmt.Print(".")

		if err := stream.Send(&pbs.StreamMsg{
			Data: chunk,
		}); err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}
	}

	fmt.Printf(" OK\n")
	_, err = stream.CloseAndRecv()
	return b, err
}

func (c *Client) CreateBuildGit(app *pb.App, version string, procfile []byte) (*pb.Build, error) {
	return c.ProviderServiceClient.BuildCreate(ctx, &pbs.CreateBuildRequest{
		App:      app.Name,
		Version:  version,
		Procfile: procfile,
	})
}

func (c *Client) GetBuilds(app string, limit int) (pb.Builds, error) {
	ret, err := c.ProviderServiceClient.BuildList(ctx, &pbs.AppListRequest{
		Name:  app,
		Limit: int64(limit),
	})

	if err != nil {
		return nil, err
	}
	return ret.Builds, nil
}

func (c *Client) GetBuild(app, id string) (*pb.Build, error) {
	return c.ProviderServiceClient.BuildGet(ctx, &pbs.AppIdRequest{App: app, Id: id})
}

func (c *Client) DeleteBuild(app, id string) error {
	_, err := c.ProviderServiceClient.BuildDelete(ctx, &pbs.AppIdRequest{App: app, Id: id})
	return err
}
