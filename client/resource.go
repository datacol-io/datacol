package client

import (
	"github.com/dinesh/datacol/client/models"
)

var rs_bucket = []byte("resources")

func (c *Client) GetResource(name string) (*models.Resource, error) {
	return c.Provider().ResourceGet(name)
}

func (c *Client) CreateResource(kind string, options map[string]string) (*models.Resource, error) {
	return c.Provider().ResourceCreate(options["name"], kind, options)
}

func (c *Client) CreateResourceLink(app, name string) error {
	_, err := c.Provider().ResourceLink(app, name)
	return err
}

func (c *Client) DeleteResourceLink(app, name string) error {
	_, err := c.Provider().ResourceUnlink(app, name)
	return err
}
