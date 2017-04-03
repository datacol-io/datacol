package main

import (
  "fmt"
  "gopkg.in/urfave/cli.v2"
  "github.com/dinesh/datacol/cmd/stdcli"
  "github.com/dinesh/datacol/client/models"
)

func init(){
  stdcli.AddCommand(cli.Command{
    Name: "deploy",
    Usage: "deploy an app in cluster",
    Action: cmdDeploy,
    Flags: []cli.Flag{
      cli.StringFlag{
        Name:   "image, i",
        Usage:  "docker image to use",
      },
      cli.IntFlag{
        Name:   "port, p",
        Usage:  "service port",
        Value:  8080,
      },
      cli.StringFlag{
        Name: "build, b",
        Usage: "Build id to use",
      },
      cli.BoolTFlag{
        Name: "wait, w",
        Usage: "Wait for the app become available",
      },
    },
  })
}

func cmdDeploy(c *cli.Context) error {
  dir, name, err := getDirApp(".")
  if err != nil { return err }

  client := getClient(c)
  app, err := client.GetApp(name)
  if err != nil { 
    return app404Err(name)
  }

  var build *models.Build
  buildId := c.String("build")

  if len(buildId) == 0 {
    build = client.NewBuild(app)
    if err = executeBuildDir(c, build, dir); err != nil {
      return err
    }
  } else {
    b, err := client.GetBuild(buildId)
    if err != nil { return err }
    if b == nil {
      return fmt.Errorf("No build found by id: %s.", buildId)
    }

    build = b
  }

  fmt.Printf("Deploying build %s\n", build.Id)
  r := client.NewRelease(build)

  port := c.Int("port")
  wait := c.BoolT("wait")
  if err = client.DeployRelease(r, port, c.String("image"), c.String("env"), wait); err != nil {
    return err
  }

  app, _ = client.GetApp(name)
  if len(app.HostPort) > 0 {
    fmt.Printf("\nDeployed at %s\n", app.HostPort)
  } else {
    fmt.Println("DONE.")
  }
  
  return nil
}
