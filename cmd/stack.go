package main

import (
  "gopkg.in/urfave/cli.v2"
  "github.com/dinesh/rz/cmd/stdcli"
)

func init(){
  stdcli.AddCommand(cli.Command{
    Name:       "init",
    UsageText:  "create new stack",
    Action:     cmdStackCreate,
    Flags:      []cli.Flag{
      &cli.StringFlag{
        Name: "stack",
        Usage: "Name of stack",
      },
      &cli.StringFlag{
        Name: "project-id",
        Usage: "GCP project-id to use",
      },
      &cli.StringFlag{
        Name: "zone",
        Usage: "zone for stack",
        Value: "us-east1-b",
      },
      &cli.StringFlag{
        Name:  "bucket",
        Usage: "GCP storage bucket",
        Value: "rzdev",
      },
      &cli.IntFlag{
        Name: "nodes",
        Usage: "number of nodes in container cluster",
        Value: 3,
      },
    },
  })

  stdcli.AddCommand(cli.Command{
    Name:      "destroy",
    UsageText: "destroy current stack",
    Action:     cmdStackDestroy,
  })
}

func cmdStackCreate(c *cli.Context) error {
  stdcli.CheckFlagsPresence(c, "stack", "project-id")

  stackName := c.String("stack")
  projectId := c.String("project-id")
  zone := c.String("zone")
  nodes := c.Int("nodes")
  bucket := c.String("bucket")

  client := getAnonClient(c)
  client.StackName = stackName

  _, err := client.CreateStack(projectId, zone, bucket, nodes)
  if err != nil {
    return err
  }

  return nil
}

func cmdStackDestroy(c *cli.Context) error {
  client := getClient(c)
  if err := client.DestroyStack(); err != nil {
    return err
  }

  return nil
  // return client.Stack.Delete()
}






