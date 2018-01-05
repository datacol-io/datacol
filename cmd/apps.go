package cmd

import (
	"errors"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go/term"
	"github.com/dinesh/datacol/cmd/stdcli"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/validation"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "apps",
		Usage:  "Manage your apps in a stack",
		Action: cmdAppsList,
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "create",
				Action: cmdAppCreate,
				Flags: []cli.Flag{
					appFlag,
					cli.StringFlag{
						Name:  "repo-url",
						Usage: "Repository url (github or codecommit)",
					},
				},
			},
			cli.Command{
				Name:   "delete",
				Action: cmdAppDelete,
				Flags:  []cli.Flag{appFlag},
			},
			cli.Command{
				Name:   "info",
				Action: cmdAppInfo,
				Flags:  []cli.Flag{},
			},
		},
	})

	stdcli.AddCommand(cli.Command{
		Name:   "restart",
		Usage:  "restart an app",
		Action: cmdAppRestart,
		Flags:  []cli.Flag{appFlag},
	})
}

func cmdAppRestart(c *cli.Context) error {
	_, app, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	stdcli.ExitOnError(client.RestartApp(app))

	term.Successln("RESTARTED")
	return nil
}

func cmdAppsList(c *cli.Context) error {
	api, close := getApiClient(c)
	defer close()

	apps, err := api.GetApps()
	stdcli.ExitOnError(err)

	if len(apps) == 0 {
		fmt.Println("No apps found.")
	} else {
		fmt.Println(toJson(apps))
	}
	return nil
}

func cmdAppCreate(c *cli.Context) error {
	name := c.String("app")

	if len(name) == 0 {
		_, n, err := getDirApp(".")
		if err != nil {
			return err
		}
		name = n
	}

	errs := validation.IsDNS1123Label(name)
	if len(errs) > 0 {
		term.Warningln(fmt.Sprintf("Invalid app name: %s", name))
		for _, e := range errs {
			log.Errorf(e)
		}
		os.Exit(1)
	}

	api, close := getApiClient(c)
	defer close()

	app, err := api.CreateApp(name, c.String("repo-url"))
	stdcli.ExitOnError(err)

	stdcli.ExitOnError(stdcli.WriteAppSetting("app", name))

	stdcli.ExitOnError(stdcli.WriteAppSetting("stack", api.Name))

	// todo: better way to hardcode stackname for app. we use <stack>-app-<name> for cloudformation.
	if api.IsAWS() {
		stdcli.ExitOnError(waitForAwsResource("app-"+name, "CREATE", api))
	}

	fmt.Printf("%s is created.\n", app.Name)
	return nil
}

func cmdAppInfo(c *cli.Context) error {
	_, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	api, close := getApiClient(c)
	defer close()

	app, err := api.GetApp(name)
	stdcli.ExitOnError(err)

	fmt.Printf("%s", toJson(app))
	return nil
}

func cmdAppDelete(c *cli.Context) error {
	abs, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	api, close := getApiClient(c)
	defer close()

	stdcli.ExitOnError(api.DeleteApp(name))

	if api.IsAWS() {
		stdcli.ExitOnError(waitForAwsResource("app-"+name, "DELETE", api))
	}

	stdcli.ExitOnError(stdcli.RmSettingDir(abs))

	fmt.Println("Done")
	return nil
}

func app404Err(name string) error {
	return errors.New(fmt.Sprintf("No app found by name: %s. Please create by running `datacol apps create`", name))
}
