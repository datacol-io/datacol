package main

import (
  "os"
  "log"
  "path/filepath"

  "github.com/dinesh/rz/client"
  "github.com/dinesh/rz/cmd/stdcli"
  "gopkg.in/urfave/cli.v2"
)

func main(){
  if len(os.Args) > 1 {
    if os.Args[1] == "kubectl" {
      cmdKubectl(os.Args[2:])
      return
    }
  }

  app := stdcli.New()
  app.Run(os.Args)
  defer closeDb()
}

func getClient(c *cli.Context) *client.Client {
  stack := stdcli.GetStack()
  conn := &client.Client{
    Version: c.App.Version,
  }

  if err := conn.SetStack(stack); err != nil {
    log.Fatal(err)
  }

  return conn
}

func getDirApp(path string) (string, string, error) {
  abs, err := filepath.Abs(path)
  if err != nil {
    return abs, "", err
  }

  app := filepath.Base(abs)
  return abs, app, nil
}

func getAnonClient(c *cli.Context) *client.Client {
  return &client.Client{}
}

func closeDb(){
  if client.DB != nil {
    client.DB.Close()
  }
}