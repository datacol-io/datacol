package client

import (
  "time"
  "fmt"
  "sort"
  "encoding/json"

  "github.com/dinesh/rz/client/models"
)

var (
  b_bucket = []byte("builds")
)

func (c *Client) NewBuild(app *models.App) *models.Build {
   b := &models.Build{
    App: app.Name,
    Id: generateId("B", 5),
    Status: "creating",
    CreatedAt: time.Now(),
  }
  
  return b
}

func (c *Client) LatestBuild(app *models.App) (*models.Build, error) {
  bbx, _ := DB.New(b_bucket)
  items, err := bbx.Items()
  if err != nil { return nil, err }

  var builds []models.Build

  for _, item := range items {
    var b models.Build
    err := json.Unmarshal(item.Value, &b)
    if err != nil { return nil, err }
    if b.Status == "success" && b.App == app.Name {
      builds = append(builds, b)
    }
  }

  sort.Slice(builds, func(i, j int) bool {
    return builds[i].CreatedAt.Second() < builds[j].CreatedAt.Second()
  })

  if len(builds) > 0 {
    return &builds[0], nil
  } else {
    return nil, fmt.Errorf("build not found")
  }
}
