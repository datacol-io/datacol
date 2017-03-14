package client

import (
  "encoding/json"
)

var (
  a_bucket = []byte("apps")
)

type App struct {
  Name    string    `json: "name"`
  Status  string    `json: "status"`
  Release string    `json: "release"`
}

type Apps []*App

func (c *Client) GetApps() (Apps, error) {
  abx, _ := DB.New(a_bucket)
  items, err := abx.Items()
  if err != nil { return nil, err }

  res := make(Apps, len(items))

  for _, item := range items {
    var a App
    if err := json.Unmarshal(item.Value, &a); err != nil {
      return nil, err
    }
    res = append(res, &a)
  }

  return res, nil
}

func (c *Client) GetApp(name string) (*App, error) {
  abx, _ := DB.New(a_bucket)
  item, err := abx.Get([]byte(name))
  if err != nil { return nil, err }
  
  var a App

  if err := json.Unmarshal(item, &a); err != nil {
    return nil, err
  }

  return &a, nil 
}

func (c *Client) CreateApp(name string) (*App, error) {
  app := &App{
    Name:   name,
    Status: "created",
  }
  
  if err := app.Persist(); err != nil {
    return nil, err
  }

  return app, nil
}

func (app *App) Persist() error {
  abx, _ := DB.New(a_bucket)
  encoded, err := json.Marshal(app)
  if err != nil { return err }

  return abx.Put([]byte(app.Name), encoded)
}

