package cmd

import (
	"fmt"

	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "env",
		Usage:  "manage environment variables for an app",
		Action: cmdConfigList,
		Flags:  []cli.Flag{&appFlag, &stackFlag},
		Subcommands: []cli.Command{
			{
				Name:      "set",
				UsageText: "set env variables",
				Action:    cmdConfigSet,
				Flags:     []cli.Flag{&appFlag, &stackFlag},
			},
			{
				Name:      "unset",
				UsageText: "unset env vars",
				Action:    cmdConfigUnset,
				Flags:     []cli.Flag{&appFlag, &stackFlag},
			},
		},
	})
}

func cmdConfigList(c *cli.Context) error {
	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	ct, close := getApiClient(c)
	defer close()

	if _, err = ct.GetApp(name); err != nil {
		err = fmt.Errorf("failed to fetch app: %v", err)
		stdcli.ExitOnError(err)
	}

	env, err := ct.GetEnvironment(name)
	stdcli.ExitOnError(err)

	data := ""
	for key, value := range env {
		data += fmt.Sprintf("%s=%s\n", key, value)
	}

	fmt.Printf(data)
	return nil
}

func cmdConfigSet(c *cli.Context) error {
	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	ct, close := getApiClient(c)
	defer close()

	env, err := ct.GetEnvironment(name)
	if err != nil {
		env = pb.Environment{}
	}

	data := ""
	for key, value := range env {
		data += fmt.Sprintf("%s=%s\n", key, value)
	}

	// handle args
	for _, value := range c.Args() {
		data += fmt.Sprintf("%s\n", value)
	}

	stdcli.ExitOnError(ct.SetEnvironment(name, data))
	return nil
}

func cmdConfigUnset(c *cli.Context) error {
	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	env, err := client.GetEnvironment(name)
	stdcli.ExitOnError(err)

	for _, key := range c.Args() {
		delete(env, key)
	}

	data := ""
	for key, value := range env {
		data += fmt.Sprintf("%s=%s\n", key, value)
	}

	stdcli.ExitOnError(client.SetEnvironment(name, data))
	return nil
}
