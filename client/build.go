package client

import (
  "time"
  "fmt"
  "encoding/json"
)

var (
  b_bucket = []byte("builds")
)

type Build struct {
  Id  string            `json: "id"`
  App string            `json: "app"`
  Status string         `json: "status"`
  CreatedAt time.Time   `json: "created_at"`
}

func (c *Client) NewBuild(app *App) *Build {
   b := &Build{
    App: app.Name,
    Id: generateId("B", 5),
    Status: "creating",
    CreatedAt: time.Now(),
  }
  
  return b
}

func (c *Client) LatestBuild(app *App) (*Build, error) {
  bbx, _ := DB.New(b_bucket)
  items, err := bbx.Items()
  if err != nil { return nil, err }

  for _, item := range items {
    var b Build
    err := json.Unmarshal(item.Value, &b)
    if err != nil { return nil, err }
    if b.Status == "success" && b.App == app.Name {
      return &b, nil
    }
  }

  return nil, fmt.Errorf("build not found")
}

func (b *Build) Persist() error {
  bbx, _ := DB.New(b_bucket)
  encoded, err := json.Marshal(b)
  if err != nil { return err }
  return bbx.Put([]byte(b.Id), encoded)
}