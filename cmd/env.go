package main

import (
	"fmt"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:   "env",
		Usage:  "manage environment variables for an app",
		Action: cmdConfigList,
		Subcommands: []*cli.Command{
			{
				Name:      "set",
				UsageText: "set env variables",
				Action:    cmdConfigSet,
			},
			{
				Name:      "unset",
				UsageText: "unset env vars",
				Action:    cmdConfigUnset,
			},
		},
	})
}

func cmdConfigList(c *cli.Context) error {
	_, name, err := getDirApp(".")
	if err != nil {
		return err
	}

	ct, close := getApiClient(c)
	defer close()

	if _, err = ct.GetApp(name); err != nil {
		return fmt.Errorf("failed to fetch app: %v", err)
	}

	env, err := ct.GetEnvironment(name)
	if err != nil {
		return err
	}

	data := ""
	for key, value := range env {
		data += fmt.Sprintf("%s=%s\n", key, value)
	}

	fmt.Printf(data)
	return nil
}

func cmdConfigSet(c *cli.Context) error {
	_, name, err := getDirApp(".")
	if err != nil {
		return err
	}

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
	for _, value := range c.Args().Slice() {
		data += fmt.Sprintf("%s\n", value)
	}

	return ct.SetEnvironment(name, data)
}

func cmdConfigUnset(c *cli.Context) error {
	_, name, err := getDirApp(".")
	if err != nil {
		return err
	}

	client, close := getApiClient(c)
	defer close()

	env, err := client.GetEnvironment(name)
	if err != nil {
		return err
	}

	keyvar := c.Args().First()
	data := ""
	for key, value := range env {
		if key != keyvar {
			data += fmt.Sprintf("%s=%s\n", key, value)
		}
	}

	return client.SetEnvironment(name, data)
}
