package main

import (
	"os"
	"time"

	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:   "logs",
		Usage:  "streams logs for an app",
		Action: cmdAppLogStream,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "follow",
				Aliases: []string{"f"},
				Usage:   "keep streaming new log output (default)",
			},
			&cli.DurationFlag{
				Name:  "since",
				Usage: "show logs since a duration (e.g. 10m or 1h2m10s)",
				Value: 2 * time.Minute,
			},
		},
	})
}

func cmdAppLogStream(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	if c.NArg() > 0 {
		name = c.Args().Get(0)
	}

	err = client.StreamAppLogs(name, c.Bool("follow"), c.Duration("since"), os.Stdout)
	stdcli.ExitOnError(err)

	return err
}
