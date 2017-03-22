package main
  
import (
  "fmt"
  "strings"
  "gopkg.in/urfave/cli.v2"
  "github.com/dinesh/datacol/cmd/stdcli"
  "github.com/dinesh/datacol/client/models"
)

func init(){
  stdcli.AddCommand(cli.Command{
    Name:  "env",
    UsageText: "manage env vars for an app",
    Action: cmdConfigList,
    Subcommands: []cli.Command{
      {
        Name:       "set",
        UsageText:  "set env variables",
        Action:     cmdConfigSet,
      },
      {
        Name:      "unset",
        UsageText: "unset env vars",
        Action:    cmdConfigUnset,
      },
    },
  })
}

func cmdConfigList(c *cli.Context) error {
  _, name, err := getDirApp(".")
  if err != nil { return err }

  env, err := getClient(c).Provider().EnvironmentGet(name)
  if err != nil {
    return err
  }

  data := ""
  for key, value := range env {
    data += fmt.Sprintf("%s=%s\n", key, value)
  }

  fmt.Printf(data)
  return nil
}

func cmdConfigSet(c *cli.Context) error {
    _, name, err := getDirApp(".")
  if err != nil { return err }

  provider := getClient(c).Provider()
  env, err := provider.EnvironmentGet(name)
  if err != nil { 
    env = models.Environment{}
  }

  data := ""
  for key, value := range env {
    data += fmt.Sprintf("%s=%s\n", key, value)
  }

  // handle args
  for _, value := range c.Args() {
    data += fmt.Sprintf("%s\n", value)
  }

  return provider.EnvironmentSet(name, strings.NewReader(data))
}

func cmdConfigUnset(c *cli.Context) error {
  _, name, err := getDirApp(".")
  if err != nil { return err }

  provider := getClient(c).Provider()
  env, err := provider.EnvironmentGet(name)
  if err != nil {
    return err
  }

  keyvar := c.Args()[0]
  data := ""
  for key, value := range env {
    if key != keyvar {
      data += fmt.Sprintf("%s=%s\n", key, value)
    }
  }

  return provider.EnvironmentSet(name, strings.NewReader(data))
}
