package main

import (
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/dinesh/datacol/client"
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
)

var (
	verbose   = false
	stackFlag cli.StringFlag
	appFlag   cli.StringFlag
)

func init() {
	verbose = os.Getenv("DATACOL_DEBUG") == "1" || os.Getenv("DATACOL_DEBUG") == "true"

	stackFlag = cli.StringFlag{
		Name:   "stack",
		Usage:  "stack name",
		EnvVar: "DATACOL_STACK,STACK",
	}

	appFlag = cli.StringFlag{
		Name:  "app, a",
		Usage: "app name inferred from current directory if not specified",
	}
}

func main() {
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

func getClient(c *cli.Context) *client.Client {
	stack := stdcli.GetStack()
	conn := &client.Client{
		Version: c.App.Version,
	}

	if err := conn.SetStack(stack); err != nil {
		log.Fatal(err)
	}

	return conn
}

func getDirApp(path string) (string, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return abs, "", err
	}

	app := stdcli.GetSetting("app")
	if len(app) == 0 {
		app = filepath.Base(abs)
	}
	return abs, app, nil
}

func getAnonClient(c *cli.Context) *client.Client {
	return &client.Client{}
}
