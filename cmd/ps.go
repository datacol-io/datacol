package cmd

import (
	"github.com/appscode/go/term"
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:   "ps",
		Usage:  "manage process in an app",
		Action: cmdAppPS,
	})

	stdcli.AddCommand(&cli.Command{
		Name:   "scale",
		Usage:  "scale the number of workers for a process",
		Action: cmdAppScale,
	})
}

func cmdAppPS(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	items, err := client.ListProcess(name)
	stdcli.ExitOnError(err)

	term.Println(toJson(items))
	return nil
}

func cmdAppScale(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	return nil
}
