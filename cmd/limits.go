package cmd

import (
	"fmt"
	"strings"

	term "github.com/appscode/go/term"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "limits",
		Usage:  "manage resource contraints",
		Action: cmdAppPS,
		Flags:  []cli.Flag{&appFlag},
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Usage:     "set the resource constraint",
				ArgsUsage: "<proctype>=<limit> or <proctype>=<limit/request>",
				Action:    cmdLimitsSet,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name: "memory",
					},
					cli.BoolFlag{
						Name: "cpu",
					},
				},
			},
			{
				Name:      "unset",
				Usage:     "unset the resource constraint",
				ArgsUsage: "<proctype>",
				Action:    cmdLimitsUnSet,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name: "memory",
					},
					cli.BoolFlag{
						Name: "cpu",
					},
				},
			},
		},
	})
}

func cmdLimitsSet(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	if c.NArg() < 1 {
		term.Warningln("No process type provided")
		stdcli.Usage(c)
	}

	resource, payload := parseResourceArgs(c, false)
	stdcli.ExitOnError(client.UpdateProcessLimits(name, resource, payload))

	term.Println("DONE")
	return nil
}

func cmdLimitsUnSet(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	if c.NArg() < 1 {
		term.Warningln("No process type provided")
		stdcli.Usage(c)
	}

	resource, payload := parseResourceArgs(c, true)
	stdcli.ExitOnError(client.UpdateProcessLimits(name, resource, payload))

	term.Println("DONE")
	return nil
}

func parseResourceArgs(c *cli.Context, unset bool) (string, map[string]string) {
	payload := make(map[string]string)

	for _, args := range c.Args() {
		parts := strings.Split(args, "=")
		if unset {
			payload[parts[0]] = "0"
		} else {
			if len(parts) < 2 {
				term.Warningln("Invalid argument fromat:", args)
				stdcli.Usage(c)
			}

			proctype, value := parts[0], parts[1]
			payload[proctype] = value
		}
	}

	isMemory := c.Bool("memory")
	isCpu := c.Bool("cpu")

	if !isMemory && !isCpu {
		isMemory = true // set the memory constraints by default
	}

	if isMemory {
		return "memory", payload
	} else if isCpu {
		return "cpu", payload

	}

	stdcli.ExitOnError(fmt.Errorf("Unsupported resource type."))

	return "", payload
}
