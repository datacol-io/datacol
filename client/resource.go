package client

import (
	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
)

func (c *Client) ListResources() (pb.Resources, error) {
	ret, err := c.ProviderServiceClient.ResourceList(ctx, &pbs.ListRequest{})
	if err != nil {
		return nil, err
	}
	return ret.Resources, nil
}

func (c *Client) GetResource(name string) (*pb.Resource, error) {
	return c.ProviderServiceClient.ResourceGet(ctx, &pbs.AppRequest{Name: name})
}

func (c *Client) CreateResource(kind string, options map[string]string) (*pb.Resource, error) {
	return c.ProviderServiceClient.ResourceCreate(ctx, &pbs.CreateResourceRequest{
		Name:   options["name"],
		Kind:   kind,
		Params: options,
	})
}

func (c *Client) DeleteResource(name string) error {
	_, err := c.ProviderServiceClient.ResourceDelete(ctx, &pbs.AppRequest{Name: name})
	return err
}

func (c *Client) CreateResourceLink(app, name string) error {
	_, err := c.ProviderServiceClient.ResourceLink(ctx, &pbs.AppResourceReq{App: app, Resource: name})
	return err
}

func (c *Client) DeleteResourceLink(app, name string) error {
	_, err := c.ProviderServiceClient.ResourceUnlink(ctx, &pbs.AppResourceReq{App: app, Resource: name})
	return err
}
