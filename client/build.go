package client

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/dinesh/datacol/client/models"
)

var (
	b_bucket = []byte("builds")
)

func (c *Client) NewBuild(app *models.App) *models.Build {
	b := &models.Build{
		App:       app.Name,
		Id:        generateId("B", 5),
		Status:    "creating",
		CreatedAt: time.Now(),
	}

	return b
}

func (c *Client) LatestBuild(name string) (*models.Build, error) {
	allbuilds, err := c.GetBuilds(name)
	if err != nil {
		return nil, err
	}

	var builds models.Builds

	for _, b := range allbuilds {
		if b.Status == "success" {
			builds = append(builds, b)
		}
	}

	sort.Slice(builds, func(i, j int) bool {
		return builds[i].CreatedAt.Second() < builds[j].CreatedAt.Second()
	})

	if len(builds) > 0 {
		return builds[0], nil
	} else {
		return nil, fmt.Errorf("build not found")
	}
}

func (c *Client) GetBuilds(app string) (models.Builds, error) {
	items, err := getList(b_bucket)
	if err != nil {
		return nil, err
	}

	var builds models.Builds

	for _, item := range items {
		var b models.Build
		err := json.Unmarshal(item.Value, &b)
		if err != nil {
			return nil, err
		}

		if b.App == app {
			builds = append(builds, &b)
		}
	}

	return builds, nil
}

func (c *Client) GetBuild(id string) (*models.Build, error) {
	item, err := getV(b_bucket, []byte(id))
	if err != nil {
		return nil, err
	}

	var b models.Build
	if err := json.Unmarshal(item, &b); err != nil {
		return nil, err
	}

	return &b, nil
}

func (c *Client) DeleteBuild(Id string) error {
	return deleteV(b_bucket, Id)
}
