package client

import (
  "time"
  "io"
  "path/filepath"
  "encoding/json"
  "github.com/dinesh/datacol/client/models"
)

var (
  a_bucket = []byte("apps")
)

func (c *Client) GetApps() (models.Apps, error) {
  abx, _ := DB.New(a_bucket)
  items, err := abx.Items()
  if err != nil { return nil, err }

  res := make(models.Apps, len(items))

  for i, item := range items {
    var a models.App
    if err := json.Unmarshal(item.Value, &a); err != nil {
      return nil, err
    }
    res[i] = &a
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
    Stack:  c.Stack.Name,
  }
  
  if err := Persist(a_bucket, app.Name, app); err != nil {
    return nil, err
  }

  return app, nil
}

func (c *Client) DeleteApp(name string) error {
  if err := c.Provider().AppDelete(name); err != nil {
    return err
  }

  abx, _ := DB.New(a_bucket)
  return abx.Delete([]byte(name))
}


func (c *Client) StreamAppLogs(name string, follow bool, since time.Duration, out io.Writer) error {
  opts := models.LogStreamOptions{Since: since, Follow: follow}
  cfgpath := filepath.Join(c.configRoot(), "kubeconfig")
  return c.Provider().LogStream(cfgpath, name, out, opts)
}



