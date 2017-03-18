package main

import (
  "time"
  "os"
  "github.com/dinesh/rz/cmd/stdcli"
  "gopkg.in/urfave/cli.v2"
)

func init(){
  stdcli.AddCommand(cli.Command{
    Name: "logs",
    UsageText: "streams logs for an app",
    Action: cmdAppLogStream,
    Flags: []cli.Flag{
      cli.BoolTFlag{
        Name:  "follow",
        Usage: "keep streaming new log output (default)",
      },
      cli.DurationFlag{
        Name:  "since",
        Usage: "show logs since a duration (e.g. 10m or 1h2m10s)",
        Value: 2 * time.Minute,
      },
    },
  })
}

func cmdAppLogStream(c *cli.Context) error {
  _, name, err := getDirApp(".")
  if err != nil { return err }

  return getClient(c).StreamAppLogs(name, c.BoolT("follow"), c.Duration("since"), os.Stdout)
}
