package client

import (
	pb "github.com/dinesh/datacol/api/models"
  pbs "github.com/dinesh/datacol/api/controller"
)

func (c *Client) GetResource(name string) (*pb.Resource, error) {
	return c.ProviderServiceClient.ResourceGet(ctx, &pbs.AppRequest{Name: name})
}

func (c *Client) CreateResource(kind string, options map[string]string) (*pb.Resource, error) {
	return c.ProviderServiceClient.ResourceCreate(ctx, &pbs.CreateResourceRequest{
    Name: options["name"], 
    Kind: kind, 
    Params: options,
  })
}

func (c *Client) CreateResourceLink(app, name string) error {
	_, err := c.ProviderServiceClient.ResourceLink(ctx, &pbs.AppResourceReq{App: app, Name: name})
	return err
}

func (c *Client) DeleteResourceLink(app, name string) error {
	_, err := c.ProviderServiceClient.ResourceUnlink(ctx, &pbs.AppResourceReq{App: app, Name: name})
	return err
}
