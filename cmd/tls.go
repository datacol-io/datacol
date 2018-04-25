package cmd

import (
	"io/ioutil"

	term "github.com/appscode/go/term"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:  "tls",
		Usage: "manage certificates for an app",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				ArgsUsage: "<domain> <tls.crt> <tls.key>",
				Action:    cmdTLSAdd,
			},
		},
	})
}

func cmdTLSAdd(c *cli.Context) error {
	if c.NArg() < 3 {
		term.Warningln("Missing required arguments")
		stdcli.Usage(c)
	}

	domain, tlsCrtPath, tlsKeyPath := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)

	crtData, err := ioutil.ReadFile(tlsCrtPath)
	if err != nil {
		stdcli.ExitOnError(err)
	}

	keyData, err := ioutil.ReadFile(tlsKeyPath)
	if err != nil {
		stdcli.ExitOnError(err)
	}

}
