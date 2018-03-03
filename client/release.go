package client

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	pbs "github.com/datacol-io/datacol/api/controller"
	"github.com/datacol-io/datacol/api/models"

	"time"
)

func (c *Client) GetReleases(app string) (models.Releases, error) {
	ret, err := c.ProviderServiceClient.ReleaseList(ctx, &pbs.AppRequest{Name: app})
	if err != nil {
		return nil, err
	}
	return ret.Releases, nil
}

func (c *Client) DeleteRelease(app, id string) error {
	_, err := c.ProviderServiceClient.ReleaseDelete(ctx, &pbs.AppIdRequest{App: app, Id: id})
	return err
}

func (c *Client) ReleaseBuild(build *models.Build, options models.ReleaseOptions) (*models.Release, error) {
	r, err := c.ProviderServiceClient.BuildRelease(ctx, &pbs.CreateReleaseRequest{
		Build:  build,
		Domain: options.Domain,
	})

	if err != nil {
		return r, err
	}

	if options.Wait {
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
