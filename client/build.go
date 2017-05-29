package client

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"io"
)

const chunkSize = 1024 * 1024 * 2

func (c *Client) CreateBuild(app *pb.App, data []byte) (*pb.Build, error) {
	r := bytes.NewReader(data)
	stream, err := c.ProviderServiceClient.BuildCreate(ctx)

	if err != nil {
		return nil, err
	}

	for {
		chunk := make([]byte, chunkSize)
		n, err := r.Read(chunk)
		if n == 0 && err == io.EOF {
			break
		}

		fmt.Print(".")
		if err := stream.Send(&pbs.CreateBuildRequest{Data: chunk, App: app.Name}); err != nil {
			if err == io.EOF {
				return nil, err
			}
		}
	}

	log.Debugf("Uploaded all chunks")
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
