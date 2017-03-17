package main

import (
  "gopkg.in/urfave/cli.v2"
  "github.com/dinesh/rz/cmd/stdcli"
  "github.com/dinesh/rz/client"
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
        Value: "dev",
      },
      &cli.StringFlag{
        Name: "project",
        Usage: "GCP project id to use",
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
      &cli.BoolFlag{
        Name: "dry-run",
        Usage: "dry run mode",
      },
      &cli.StringFlag{
        Name: "cluster",
        Usage: "IP:[PORT] for existing Kuberenetes cluster",
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
  stdcli.CheckFlagsPresence(c, "project")

  stackName := c.String("stack")
  projectId := c.String("project")
  zone := c.String("zone")
  nodes := c.Int("nodes")
  bucket := c.String("bucket")
  dryRun := c.Bool("dry-run")
  cluster := c.String("cluster")

  ac := getAnonClient(c)
  st, err := client.FindStack(stackName)

  if err != nil {
    ac.StackName = stackName
    if st, err = ac.CreateStack(projectId, zone, bucket); err != nil {
      return err
    }
  }

  if dryRun { return nil }
  
  ac.SetStack(st.Name)
  
  return ac.DeployStack(st, cluster, nodes)
}

func cmdStackDestroy(c *cli.Context) error {
  client := getClient(c)

  if err := client.DestroyStack(); err != nil {
    return err
  }

  return client.Stack.Delete()
}

