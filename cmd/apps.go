package main

import (
  "os"
  "fmt"
  "errors"
  "gopkg.in/urfave/cli.v2"
  log "github.com/Sirupsen/logrus"
  "github.com/dinesh/datacol/cmd/stdcli"
  "k8s.io/client-go/pkg/util/validation"
)

func init(){
  stdcli.AddCommand(cli.Command{
    Name:       "apps",
    Usage:      "Manage your apps in a stack",
    Action:     cmdAppsList,
    Subcommands: []cli.Command{
      cli.Command{
        Name:   "create",
        Action: cmdAppCreate,
        Flags: []cli.Flag{appFlag},
      },
      cli.Command{
        Name:    "delete",
        Action:  cmdAppDelete,
      },
      cli.Command{
        Name:   "info",
        Action: cmdAppInfo,
        Flags:  []cli.Flag{
          cli.BoolFlag{
            Name:    "wait, w",
            Usage:   "wait for GCP to allocate IP",
          },
        },
      },
    },
  })

  stdcli.AddCommand(cli.Command{
    Name:     "restart",
    Usage:    "restart an app",
    Action:   cmdAppRestart,
    Flags:    []cli.Flag{appFlag},
  })
}

func cmdAppRestart(c *cli.Context) error {
    _, app, err := getDirApp(".")
  if err != nil { return err }

  if err := getClient(c).Provider().AppRestart(app); err != nil {
    return err
  }

  fmt.Println("Restarted")
  return nil
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
  name := c.String("app")
  if len(name) == 0 {
    _, n, err := getDirApp(".")
    if err != nil { return err }
    name = n
  }

  errs := validation.IsDNS1123Label(name)
  if len(errs) > 0 {
    fmt.Printf("Invalid app name: %s\n", name)
    for _, e := range errs {
      log.Errorf(e)
    }
    os.Exit(1)
  }

  client := getClient(c)
  app, err := client.CreateApp(name)
  if err != nil { return err }

  if err = stdcli.WriteSetting("app", name); err != nil {
    return err
  }

  if err = stdcli.WriteSetting("stack", client.Stack.Name); err != nil {
    return err
  }

  fmt.Printf("%s is created.\n", app.Name)
  return nil
}

func cmdAppInfo(c *cli.Context) error {
  _, name, err := getDirApp(".")
  if err != nil { return err }

  app, err := getClient(c).GetApp(name)
  if err != nil { return err }

  wait := c.Bool("wait")
  if err = getClient(c).SyncApp(app, wait); err != nil {
    return err
  }

  fmt.Printf("%+v", app)
  return nil
}

func cmdAppDelete(c *cli.Context) error {
  abs, name, err := getDirApp(".")    
  if err != nil { return err }

  app, err := getClient(c).GetApp(name)
  if err != nil { return err }

  if err = getClient(c).DeleteApp(app.Name); err != nil {
    return err
  }

  return stdcli.RmSettingDir(abs)
}

func app404Err(name string) error {
  return errors.New(fmt.Sprintf("No app found by name: %s. Please create by running $ datacol apps create", name))
}