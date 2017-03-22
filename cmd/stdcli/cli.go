package stdcli

import (
  "os"
  "fmt"
  "runtime"
  "strings"
  "io/ioutil"
  "path/filepath"

  "gopkg.in/urfave/cli.v2"
  rollbarAPI "github.com/stvp/rollbar"
)

var (
  Binary string
  Version string
  Commands []cli.Command
  localappdir string
)

func init() {
  Version  = "0.0.1-alpha"
  localappdir = ".dtcol"
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
    Error(fmt.Errorf("no stack found, Please run $] dtcol init"))
  }
  return stack
}

func AddCommand(cmd cli.Command) {
  Commands = append(Commands, cmd)
}

func GetSetting(setting string) string {
  value, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", localappdir, setting))
  if err != nil {
    return ""
  }
  output := strings.TrimSpace(string(value))

  return output
}

func WriteSetting(setting, value string) error {
  if err := os.MkdirAll(localappdir, 0777); err != nil { 
    return err
  }

  return ioutil.WriteFile(
    fmt.Sprintf(localappdir + "/%s", setting), 
    []byte(value),
    0777,
  )
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

func HandlePanicErr(err error) {
  fmt.Println(err.Error())
  rollbar(err, "error")
}

func rollbar(err error, level string) {
  if os.Getenv("TESTING") == "1" {
    return
  }

  rollbarAPI.Platform = "client"
  rollbarAPI.Token = "915b990fdfee4bd4a8c280b3a838205d"  
  var cmd string

  if len(os.Args) > 1 {
    cmd = os.Args[1]
  }

  fields := []*rollbarAPI.Field{
    {"version", Version},
    {"os", runtime.GOOS},
    {"arch", runtime.GOARCH},
    {"command", cmd},
  }

  rollbarAPI.Error(level, err, fields...)
  rollbarAPI.Wait()
}