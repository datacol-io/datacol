package cmd

import (
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/dinesh/datacol/client"
	"github.com/dinesh/datacol/cmd/stdcli"
	"github.com/dinesh/datacol/go/env"
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

func Initialize() {
	defer handlePanic()

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	app := stdcli.New()

	app.Run(os.Args)
}

func getApiClient(c *cli.Context) (*client.Client, func() error) {
	return client.NewClient(c.App.Version)
}

func getDirApp(path string) (string, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return abs, "", err
	}

	app := stdcli.GetAppSetting("app")
	if len(app) == 0 {
		app = filepath.Base(abs)
	}
	return abs, app, nil
}
