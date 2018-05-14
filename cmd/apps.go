package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"

	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go/term"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/validation"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "apps",
		Usage:  "Manage your apps in a stack",
		Action: cmdAppsList,
		Flags:  []cli.Flag{&stackFlag},
		Subcommands: []cli.Command{
			cli.Command{
				Name:   "create",
				Action: cmdAppCreate,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "Appliction name (alphanumeric)",
					},
					cli.StringFlag{
						Name:  "repo-url",
						Usage: "Repository url (github or codecommit)",
					},
				},
			},
			cli.Command{
				Name:   "delete",
				Action: cmdAppDelete,
				Flags:  []cli.Flag{&appFlag},
			},
			cli.Command{
				Name:   "info",
				Action: cmdAppInfo,
				Flags:  []cli.Flag{&appFlag},
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
	_, app, err := getDirApp(".", c)
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
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"NAME", "BUILD", "RELEASE", "DOMAINS"})
		for _, a := range apps {
			table.Append([]string{a.Name, a.BuildId, a.ReleaseId,
				strings.Join(a.Domains, "\n"),
			})
		}
		table.Render()
	}
	return nil
}

func cmdAppCreate(c *cli.Context) error {
	name := c.String("name")

	if len(name) == 0 {
		n, err := getCurrentApp(c)
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
	name, err := getCurrentApp(c)
	stdcli.ExitOnError(err)

	api, close := getApiClient(c)
	defer close()

	app, err := api.GetApp(name)
	stdcli.ExitOnError(err)

	fmt.Printf("%s", toJson(app))
	return nil
}

func cmdAppDelete(c *cli.Context) error {
	abs, name, err := getDirApp(".", c)
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
	return fmt.Errorf("No app found by name: %s. Please create by running `datacol apps create`", name)
}
