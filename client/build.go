package client

import (
	"fmt"
	"io"
	"math"

	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"google.golang.org/grpc/metadata"
)

const chunkSize = 1024 * 1024 * 1

func (c *Client) CreateBuild(app *pb.App, data []byte, procfile map[string]string) (*pb.Build, error) {
	newctx := metadata.NewOutgoingContext(ctx, metadata.Join(
		metadata.Pairs("app", app.Name),
		metadata.New(procfile),
	))

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
		size := intMin(chunkSize, len(chunk))

		if err := stream.Send(&pbs.CreateBuildRequest{
			Data: chunk,
			Size: int32(size),
			App:  app.Name,
		}); err != nil {
			if err == io.EOF {
				return nil, err
			}
		}
	}

	fmt.Printf(" OK\n")
	return stream.CloseAndRecv()
}

func (c *Client) CreateBuildGit(app *pb.App, version string) (*pb.Build, error) {
	return c.ProviderServiceClient.BuildCreate(ctx, &pbs.CreateBuildRequest{
		App:     app.Name,
		Version: version,
	})
}

func (c *Client) GetBuilds(app string) (pb.Builds, error) {
	ret, err := c.ProviderServiceClient.BuildList(ctx, &pbs.AppRequest{Name: app})
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
