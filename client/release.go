package client

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/dinesh/datacol/api/models"
	"time"
)

func (c *Client) GetReleases(app string) (models.Releases, error) {
	return c.Provider().ReleaseList(app, 20)
}

func (c *Client) DeleteRelease(app, Id string) error {
	return c.Provider().ReleaseDelete(app, Id)
}

func (c *Client) ReleaseBuild(build *models.Build, wait bool) (*models.Release, error) {
	r, err := c.ProviderServiceClient.BuildRelease(ctx, build)
	if err != nil {
		return r, err
	}

	if wait {
		err = c.waitForLoadBalancerIp(build.App)
	}

	return r, err
}

func (c *Client) waitForLoadBalancerIp(name string) error {
	waitTill := time.Now().Add(time.Duration(1) * time.Minute)
	fmt.Printf("waiting for ip")

	for {
		time.Sleep(1 * time.Second)
		fmt.Print(".")
		
		app, err := c.GetApp(name)
		if err != nil {
			return err
		}

		if len(app.Endpoint) > 0 {
			break
		}

		if time.Now().After(waitTill) {
			log.Warn("timed out. Skipping.")
			break
		}
	}

	return nil
}
