package client

import (
	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
)

func (c *Client) CreateBuild(app *pb.App, data []byte) (*pb.Build, error) {
	return c.ProviderServiceClient.BuildCreate(ctx, &pbs.CreateBuildRequest{
		App:  app.Name,
		Data: data,
	})
}

func (c *Client) GetBuilds(app string) (pb.Builds, error) {
	return c.Provider().BuildList(app, 20)
}

func (c *Client) GetBuild(app, id string) (*pb.Build, error) {
	return c.Provider().BuildGet(app, id)
}

func (c *Client) DeleteBuild(app, id string) error {
	return c.Provider().BuildDelete(app, id)
}
