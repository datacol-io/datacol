package client

import ()

type App struct {
  Name    string
  Status  string
  Release string
}

type Apps []*App

func (c *Client) GetApp(name string) (*App, error) {
  return &App{Name: name}, nil
}
