package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/dinesh/datacol/client/models"
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:   "deploy",
		Usage:  "deploy an app",
		Action: cmdDeploy,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "image, i",
				Usage: "docker image to use",
			},
			&cli.IntFlag{
				Name:  "port, p",
				Usage: "service port",
				Value: 8080,
			},
			&cli.StringFlag{
				Name:  "build, b",
				Usage: "Build id to use",
			},
			&cli.BoolFlag{
				Name:  "wait, w",
				Usage: "Wait for the app become available",
				Value: true,
			},
			&cli.StringFlag{
				Name:  "file, f",
				Usage: "path of Dockerfile or app.yaml",
			},
		},
	})
}

func cmdDeploy(c *cli.Context) error {
	dir, name, err := getDirApp(".")
	if err != nil {
		return err
	}

	client := getClient(c)
	app, err := client.GetApp(name)
	if err != nil {
		log.Warn(err)
		return app404Err(name)
	}

	var build *models.Build
	buildId := c.String("build")

	if len(buildId) == 0 {
		build = client.NewBuild(app)
		if err = executeBuildDir(c, build, dir); err != nil {
			return err
		}
	} else {
		b, err := client.GetBuild(name, buildId)
		if err != nil {
			return err
		}
		if b == nil {
			return fmt.Errorf("No build found by id: %s.", buildId)
		}

		build = b
	}

	fmt.Printf("Deploying build %s\n", build.Id)
	_, err = client.BuildRelease(build, c.Bool("wait"))

	app, _ = client.GetApp(name)
	if len(app.HostPort) > 0 {
		fmt.Printf("\nDeployed at %s\n", app.HostPort)
	} else {
		fmt.Println("DONE.")
	}

	return nil
}
