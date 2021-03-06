package cmd

import (
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/datacol-io/datacol/client"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/datacol-io/datacol/go/env"
	"github.com/urfave/cli"
)

var (
	verbose   = false
	stackFlag cli.StringFlag
	appFlag   cli.StringFlag
	ev        env.Environment
)

func init() {
	ev = env.FromHost()
	verbose = ev.DebugEnabled()

	stackFlag = cli.StringFlag{
		Name:   "stack",
		Usage:  "stack name",
		EnvVar: "STACK",
	}

	appFlag = cli.StringFlag{
		Name:  "app, a",
		Usage: "app name inferred from current directory if not specified",
	}
}

func Initialize(version, rbToken string) {
	defer handlePanic(version, rbToken)

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	app := stdcli.New()
	app.Version = version

	app.Run(os.Args)
}

func getApiClient(c *cli.Context) (*client.Client, func() error) {
	return client.NewClient(c)
}

func getDirApp(path string, c *cli.Context) (string, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return abs, "", err
	}

	app := c.String("app")
	if app == "" {
		app = stdcli.GetAppSetting("app")
	}

	if len(app) == 0 {
		app = filepath.Base(abs)
	}
	return abs, app, nil
}

// getCurrentApp support -a/--app flag for various subcommand for an APP
func getCurrentApp(c *cli.Context) (string, error) {
	if app := c.String("app"); app != "" {
		return app, nil
	}

	_, app, err := getDirApp(".", c)

	return app, err
}
