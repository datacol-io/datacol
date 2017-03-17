package main

import (
  "fmt"
  "github.com/dinesh/rz/cmd/stdcli"
  "gopkg.in/urfave/cli.v2"
)

func init(){
  stdcli.AddCommand(cli.Command{
    Name: "deploy",
    UsageText: "deploy an app",
    Action: cmdDeploy,
    Flags: []cli.Flag{
      cli.StringFlag{
        Name:   "image, i",
        Usage:  "docker image to use",
      },
      cli.IntFlag{
        Name:   "port, p",
        Usage:  "service port",
      },
    },
  })
}

func cmdDeploy(c *cli.Context) error {
  dir := "."

  if c.NArg() > 0 {
    dir = c.Args().Get(0)
  }

  client := getClient(c)
  _, name, err := getDirApp(dir)
  if err != nil { return err }

  app, err := client.GetApp(name)
  if err != nil { return err }
  
  build, err := client.LatestBuild(app)
  if err != nil { return err }

  fmt.Printf("Deploying build %s", build.Id)
  r := client.NewRelease(build)

  port := c.Int("port")
  if port == 0 { port = 80 }

  return client.DeployRelease(r, port, c.String("image"), c.String("env"))
}
