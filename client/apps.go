package client

import (
  "time"
  "io"
  "encoding/json"
  "github.com/dinesh/datacol/client/models"
)

var (
  a_bucket = []byte("apps")
)

func (c *Client) GetApps() (models.Apps, error) {
  items, err := getList(a_bucket)
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
  item, err := getV(a_bucket, []byte(name))

  if err != nil { return nil, err }
  
  var a models.App

  if err := json.Unmarshal(item, &a); err != nil {
    return nil, err
  }

  return &a, nil
}

func (c *Client) SyncApp(app *models.App, wait bool) error {
  if len(app.HostPort) == 0 && app.Status == "Running" && wait {
    if napp, _ := c.Provider().AppGet(app.Name); napp != nil {
      app.HostPort = napp.HostPort
    }
    
    if err := Persist(a_bucket, app.Name, &app); err != nil {
      return err
    }
  }
  return nil
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

  return deleteV(a_bucket, name)
}


func (c *Client) StreamAppLogs(name string, follow bool, since time.Duration, out io.Writer) error {
  opts := models.LogStreamOptions{Since: since, Follow: follow}
  return c.Provider().LogStream(name, out, opts)
}



