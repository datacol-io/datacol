package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/appscode/go/term"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "ps",
		Usage:  "manage process in an app",
		Action: cmdAppPS,
		Flags:  []cli.Flag{&appFlag},
		Subcommands: []cli.Command{
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

	stdcli.AddCommand(cli.Command{
		Name:      "scale",
		Usage:     "scale processes in an app",
		Action:    cmdAppScale,
		ArgsUsage: "<proctype>:<int> ...",
		Flags:     []cli.Flag{&appFlag},
	})
}

func cmdAppPS(c *cli.Context) error {
	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	_, err = client.GetApp(name)
	stdcli.ExitOnError(err)

	items, err := client.ListProcess(name)
	stdcli.ExitOnError(err)

	if len(items) > 0 {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetColWidth(100)
		table.SetHeader([]string{"PROCESS", "REPLICAS", "STATUS", "CPU", "MEMORY", "COMMAND"})
		for _, item := range items {
			table.Append([]string{item.Proctype,
				fmt.Sprintf("%d", item.Count),
				item.Status,
				item.Cpu,
				item.Memory,
				strings.Join(item.Command, " ")})
		}
		table.Render()
	} else {
		term.Println("No process running")
	}
	return nil
}

func cmdAppScale(c *cli.Context) error {
	name, err := getCurrentApp(c)
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
	name, err := getCurrentApp(c)
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
	name, err := getCurrentApp(c)
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
