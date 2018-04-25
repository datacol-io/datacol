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
				Usage:     "update the constraint",
				ArgsUsage: "<proctype>=<limit> or <proctype>=<limit/request>",
				Action:    cmdLimitsSet,
				Flags: []cli.Flag{
					cli.BoolTFlag{
						Name: "memory",
					},
					cli.BoolTFlag{
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

	var payload map[string]string

	for _, args := range c.Args() {
		parts := strings.Split(args, "=")
		if len(parts) < 2 {
			stdcli.ExitOnError(fmt.Errorf("Invalid argument: %v for resource constraints", args))
		}
		proctype, value := parts[0], parts[1]
		payload[proctype] = value
	}

	isMemory := c.BoolT("memory")
	isCpu := c.Bool("cpu")

	if isMemory {
		client.UpdateProcessLimits(name, "memory", payload)
	} else if isCpu {
		client.UpdateProcessLimits(name, "cpu", payload)
	} else {
		stdcli.ExitOnError("Unsupported resource type.")
	}

	return nil
}
