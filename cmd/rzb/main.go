package main

import (
  "os"
  "fmt"

  "gopkg.in/urfave/cli.v2"
  "github.com/dinesh/rz/client"
)

func main(){
  builder := func(c *cli.Context) error {
    var appDir string
    if c.NArg() > 0 {
      appDir = c.Args().Get(0)
    } else {
      dir, err := os.Getwd()
      if err != nil {
        return cli.Exit(err.Error(), 1)
      }
      appDir = dir
    }

    projectId := c.String("project-id")
    bucket := c.String("bucket")

    if appDir == "" {
      return cli.Exit("missing project directory", 1)
    }

    if projectId == "" {
      return cli.Exit("missing required option: --project-id", 1)
    }

    if bucket == "" {
      return cli.Exit("missing required option: --bucket", 1)
    }

    fmt.Printf("Building from %s\n", appDir)
    return client.ExecuteBuildDir(projectId, bucket, appDir)
  }

  onInit := func(c *cli.Context) error {
    zone := c.String("zone")
    projectId := c.String("project-id")
    rackName := c.String("rack")

    if zone == "" {
      return cli.Exit("missing required option: --zone", 1)
    }
    if projectId == "" {
      return cli.Exit("missing required option: --project-id", 1)
    }

    client.CreateStack(rackName, zone, projectId)
    return nil
  }

  projectFlag := &cli.StringFlag{Name: "project-id"}

  app := &cli.App {
    Commands: []*cli.Command{
      &cli.Command{
        Name: "build",
        UsageText: "build an app",
        Action: builder,
        Flags: []cli.Flag { 
          &cli.StringFlag{Name: "bucket"},
          projectFlag,
        },
      },
      &cli.Command{
        Name: "init",
        UsageText: "create a new stack",
        Action: onInit,
        Flags: []cli.Flag {
          &cli.StringFlag{Name: "rack"},
          &cli.StringFlag{Name: "zone" },
          projectFlag,
        },
      },
    },
  }

  app.Run(os.Args)
}