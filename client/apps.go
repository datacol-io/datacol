package client

import (
  "encoding/json"
  "github.com/dinesh/rz/client/models"
)

var (
  a_bucket = []byte("apps")
)


func (c *Client) GetApps() (models.Apps, error) {
  abx, _ := DB.New(a_bucket)
  items, err := abx.Items()
  if err != nil { return nil, err }

  res := make(models.Apps, len(items))

  for _, item := range items {
    var a models.App
    if err := json.Unmarshal(item.Value, &a); err != nil {
      return nil, err
    }
    res = append(res, &a)
  }

  return res, nil
}

func (c *Client) GetApp(name string) (*models.App, error) {
  abx, _ := DB.New(a_bucket)
  item, err := abx.Get([]byte(name))
  if err != nil { return nil, err }
  
  var a models.App

  if err := json.Unmarshal(item, &a); err != nil {
    return nil, err
  }

  return &a, nil 
}

func (c *Client) CreateApp(name string) (*models.App, error) {
  app := &models.App{
    Name:   name,
    Status: "created",
  }
  
  if err := Persist(a_bucket, app.Name, app); err != nil {
    return nil, err
  }

  return app, nil
}

func (c *Client) DeleteApp(name string) error {
  abx, _ := DB.New(a_bucket)
  return abx.Delete([]byte(name))
}

