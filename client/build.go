package client

import (
	"fmt"
	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"io"
)

const chunkSize = 1024 * 1024 * 1

func (c *Client) CreateBuild(app *pb.App, data []byte) (*pb.Build, error) {
	stream, err := c.ProviderServiceClient.BuildCreate(ctx)
	defer stream.CloseSend()

	if err != nil {
		return nil, err
	}

	numChunks := len(data)/chunkSize + 1
	fmt.Print("Uploading source ")

	for i := 0; i < numChunks; i++ {
		maxEnd := intMin((i+1)*chunkSize, len(data[i*chunkSize:]))

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
