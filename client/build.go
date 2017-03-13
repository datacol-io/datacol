package client

import (
  "time"
)

type Build struct {
  Id string       `json:"id"`
  AppName string  `json:"app"`
  *App            `json:"-"`
  Status string   `json:"status"`
  CreatedAt time.Time `json: "created_at"`
}

func (c *Client) NewBuild(app *App) *Build {
   b := &Build{
    App: app,
    AppName: app.Name,
    Id: generateId("B", 5),
    CreatedAt: time.Now(),
  }
  
  return b
}
