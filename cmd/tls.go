package cmd

import (
	"encoding/base64"
	"io/ioutil"

	term "github.com/appscode/go/term"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:  "certs",
		Usage: "manage certificates",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Usage:     "add SSL certificates for a domain",
				ArgsUsage: "<domain> <tls.crt> <tls.key>",
				Action:    cmdTLSAdd,
			},
			{
				Name:      "unset",
				Usage:     "delete SSL certificates for a domain",
				ArgsUsage: "<domain>",
				Action:    cmdTLSDelete,
			},
		},
	})
}

func cmdTLSAdd(c *cli.Context) error {
	if c.NArg() < 3 {
		term.Warningln("Missing required arguments")
		stdcli.Usage(c)
	}

	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	domain, tlsCrtPath, tlsKeyPath := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)

	crtData, err := ioutil.ReadFile(tlsCrtPath)
	if err != nil {
		stdcli.ExitOnError(err)
	}

	keyData, err := ioutil.ReadFile(tlsKeyPath)
	if err != nil {
		stdcli.ExitOnError(err)
	}

	err = client.CreateCertificate(name, domain, encodeStr(crtData), encodeStr(keyData))
	stdcli.ExitOnError(err)
	return nil
}

func cmdTLSDelete(c *cli.Context) error {
	if c.NArg() < 1 {
		term.Warningln("Missing required arguments")
		stdcli.Usage(c)
	}

	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	domain := c.Args().Get(0)

	err = client.DeleteCertificate(name, domain)
	stdcli.ExitOnError(err)
	return nil
}

func encodeStr(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
