package cmd

import (
	"strings"

	term "github.com/appscode/go/term"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:      "switch",
		Usage:     "switch between different stacks/environments",
		ArgsUsage: "<stack>[/<app>]",
		Action:    cmdSwitch,
	})
}

func cmdSwitch(c *cli.Context) error {
	if c.NArg() < 1 {
		term.Warningln("Missing required argument.")
		stdcli.Usage(c)
	}

	parts := strings.Split(c.Args().First(), "/")
	var stack, app string
	stack = parts[0]
	if len(parts) > 1 {
		app = parts[1]
	}

	stdcli.ExitOnError(stdcli.WriteAppSetting("stack", stack))
	if app != "" {
		stdcli.ExitOnError(stdcli.WriteAppSetting("app", app))
	}

	return nil
}
