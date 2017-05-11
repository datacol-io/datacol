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
	if err != nil {
		return err
	}

	client, close := getApiClient(c)
	defer close()

	if _, err := client.GetApp(name); err != nil {
		return err
	}

	args := c.Args().Slice()
	ret, err := client.RunProcess(name, args)
	if err != nil {
		return err
	}
	onApiExec(ret, args)

	return nil
}
