package client

import (
	models "github.com/dinesh/datacol/api/models"
	"io"
	"time"
)

func (c *Client) GetApps() (models.Apps, error) {
	return c.Provider().AppList()
}

func (c *Client) GetApp(name string) (*models.App, error) {
	return c.Provider().AppGet(name)
}

func (c *Client) CreateApp(name string) (*models.App, error) {
	return c.Provider().AppCreate(name)
}

func (c *Client) DeleteApp(name string) error {
	return c.Provider().AppDelete(name)
}

func (c *Client) StreamAppLogs(name string, follow bool, since time.Duration, out io.Writer) error {
	opts := models.LogStreamOptions{Since: since, Follow: follow}
	return c.Provider().LogStream(name, out, opts)
}
