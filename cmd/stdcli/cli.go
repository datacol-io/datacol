package stdcli

import (
  "path/filepath"
  "os"
  "fmt"
  "io/ioutil"
  "strings"

  "gopkg.in/urfave/cli.v2"
)

var (
  Binary string
  Version string
  Commands []cli.Command
)

func init() {
  Version  = "0.0.1-alpha"
  Binary   = filepath.Base(os.Args[0])
  Commands = []cli.Command{}
}

func New() *cli.App {
  app := &cli.App{
    Name:     Binary,
    Commands: Commands,
    Version:  Version,
    Flags:    []cli.Flag{
      cli.StringFlag{
        Name:  "app, a",
        Usage: "app name inferred from current directory if not specified",
      },
      cli.StringFlag{
        Name:  "stack",
        Usage: "stack name",
      },
    },
  }

  app.CommandNotFound = func(c *cli.Context, cmd string) {
    fmt.Fprintf(os.Stderr, "No such command \"%s\". Try `%s help`\n", cmd, Binary)
    os.Exit(1)
  }

  return app
}

func GetStack() string {
  stack := GetSetting("stack") 
  if stack == "" {
    stack = os.Getenv("STACK")
  }

  if stack == "" {
    Error(fmt.Errorf("no stack found, Please run $] rz init"))
  }
  return stack
}

func AddCommand(cmd cli.Command) {
  Commands = append(Commands, cmd)
}

func GetSetting(setting string) string {
  value, err := ioutil.ReadFile(fmt.Sprintf(".rz/%s", setting))
  if err != nil {
    return ""
  }
  output := strings.TrimSpace(string(value))

  return output
}

func WriteSetting(setting, value string) error {
  err := ioutil.WriteFile(fmt.Sprintf(".rz/%s", setting), []byte(value), 0777)

  return err
}

func CheckFlagsPresence(c *cli.Context, flags ...string){
  for _, name := range flags {
    value := c.String(name)
    if value == "" {
      Error(fmt.Errorf("Missing required param %v", name))
    }
  }
}

func Error(err error){
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }
}