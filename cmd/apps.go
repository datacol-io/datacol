package main

import (
  "fmt"
  "errors"
  "gopkg.in/urfave/cli.v2"
  "github.com/dinesh/datacol/cmd/stdcli"
)

func init(){
  stdcli.AddCommand(cli.Command{
    Name:       "apps",
    UsageText:  "Manage your apps in a stack",
    Action:     cmdAppsList,
    Subcommands: []cli.Command{
      cli.Command{
        Name:   "create",
        Action: cmdAppCreate,
      },
      cli.Command{
        Name:    "delete",
        Action:  cmdAppDelete,
      },
      cli.Command{
        Name:   "info",
        Action: cmdAppInfo,
      },
    },
  })
}

func cmdAppsList(c *cli.Context) error {
  apps, err := getClient(c).GetApps()
  if err != nil {
    return err
  }

  for _, a := range apps {
    fmt.Printf("%+v\n", a)
  }

  return nil
}

func cmdAppCreate(c *cli.Context) error {
  _, name, err := getDirApp(".")
  if err != nil { return err }
  client := getClient(c)
  
  app, err := client.CreateApp(name)
  if err != nil { return err }

  fmt.Printf("%+v created.\n", app)

  return stdcli.WriteSetting("stack", client.Stack.Name)
}

func cmdAppInfo(c *cli.Context) error {
  _, name, err := getDirApp(".")
  if err != nil { return err }

  app, err := getClient(c).GetApp(name)
  if err != nil { return err }

  fmt.Printf("%+v", app)
  return nil
}

func cmdAppDelete(c *cli.Context) error {
    _, name, err := getDirApp(".")
  if err != nil { return err }

  return getClient(c).DeleteApp(name)
}

func app404Err(name string) error {
  return errors.New(fmt.Sprintf("No app found by name: %s. Please create by running $ dcol apps create", name))
}