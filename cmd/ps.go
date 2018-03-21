package cmd

import (
	"errors"
	"fmt"

	"github.com/appscode/go/term"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "ps",
		Usage:  "manage process in an app",
		Action: cmdAppPS,
		Flags:  []cli.Flag{appFlag},
		Subcommands: []cli.Command{
			{
				Name:   "scale",
				Usage:  "scale the process",
				Action: cmdAppScale,
			},
			{
				Name:   "start",
				Usage:  "start a process",
				Action: cmdAppStart,
			},
			{
				Name:   "stop",
				Usage:  "stop a process",
				Action: cmdAppStop,
			},
		},
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

	if len(items) > 0 {
		term.Println(toJson(items))
	} else {
		term.Println("No process running")
	}
	return nil
}

func cmdAppScale(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	if err = client.SaveProcess(name, stdcli.FlagsToOptions(c, c.Args())); err != nil {
		stdcli.ExitOnError(err)
	}

	fmt.Println("OK")
	return nil
}

func cmdAppStart(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	if c.NArg() < 1 {
		stdcli.ExitOnError(errors.New("No process given"))
	}

	proc := c.Args()[0]
	if err = client.SaveProcess(name, map[string]string{proc: "1"}); err != nil {
		stdcli.ExitOnError(err)
	}

	fmt.Println("OK")
	return nil
}

func cmdAppStop(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	if c.NArg() < 1 {
		stdcli.ExitOnError(errors.New("No process given"))
	}

	proc := c.Args()[0]
	if err = client.SaveProcess(name, map[string]string{proc: "0"}); err != nil {
		stdcli.ExitOnError(err)
	}

	fmt.Println("OK")
	return nil
}
