package main

import (
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
	"os"
	"time"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "logs",
		Usage:  "streams logs for an app",
		Action: cmdAppLogStream,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "follow, f",
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
	if err != nil {
		return err
	}

	if _, err := getClient(c).GetApp(name); err != nil {
		return err
	}

	if c.NArg() > 0 {
		name = c.Args().Get(0)
	}

	return getClient(c).StreamAppLogs(name, c.Bool("follow"), c.Duration("since"), os.Stdout)
}
