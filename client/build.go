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
