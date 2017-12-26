package main

import (
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:   "run",
		Usage:  "execute a command in an app",
		Action: cmdAppRun,
	})
}

func cmdAppRun(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	args := c.Args().Slice()

	stdcli.ExitOnError(client.RunProcess(name, args))

	return nil
}
