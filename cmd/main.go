package main

import (
  "os"
  "path/filepath"

  log "github.com/Sirupsen/logrus"
  "github.com/dinesh/datacol/client"
  "github.com/dinesh/datacol/cmd/stdcli"
  "gopkg.in/urfave/cli.v2"
)

var verbose = false

func init(){
  verbose = os.Getenv("DEBUG") == "1" || os.Getenv("DEBUG") == "true"
}

func main(){
  defer handlePanic()

  log.SetFormatter(&log.TextFormatter{
    DisableTimestamp: true,
  })

  if verbose {
    log.SetLevel(log.DebugLevel)
  }

  if len(os.Args) > 0 && os.Args[1] == "kubectl" {
    cmdKubectl(os.Args[2:])
    return
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
