package cmd

import (
	"fmt"

	term "github.com/appscode/go/term"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "domains",
		Usage:  "Manage your domains for an app",
		Action: cmdDomainsList,
		Flags:  []cli.Flag{&stackFlag},
		Subcommands: []cli.Command{
			{
				Name:      "add",
				ArgsUsage: "<domain>",
				Action:    cmdAddDomain,
				Flags:     []cli.Flag{appFlag},
			},
			{
				Name:      "remove",
				ArgsUsage: "<domain>",
				Action:    cmdRemoveDomain,
				Flags:     []cli.Flag{appFlag},
			},
		},
	})
}

func cmdDomainsList(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	api, close := getApiClient(c)
	defer close()

	app, err := api.GetApp(name)
	stdcli.ExitOnError(err)

	fmt.Println(toJson(app.Domains))
	return nil
}

func cmdAddDomain(c *cli.Context) error {
	stdcli.ExitOnError(modifyDomain(c, false))
	term.Infoln("DONE")
	return nil
}

func cmdRemoveDomain(c *cli.Context) error {
	stdcli.ExitOnError(modifyDomain(c, true))
	term.Infoln("DONE")
	return nil
}

func modifyDomain(c *cli.Context, delete bool) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	api, close := getApiClient(c)
	defer close()

	if c.NArg() < 1 {
		term.Warningln("No domain provided.")
		stdcli.Usage(c)
	}

	domain := c.Args().First()
	if delete {
		domain = fmt.Sprintf(":%s", domain)
	}

	return api.AppDomainUpdate(name, domain)
}
