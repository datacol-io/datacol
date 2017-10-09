package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
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
				Name:    "image",
				Aliases: []string{"i"},
				Usage:   "docker image to use",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "service port",
				Value:   8080,
			},
			&cli.StringFlag{
				Name:    "build",
				Aliases: []string{"b"},
				Usage:   "Build id to use",
			},
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait for the app become available",
				Value:   true,
			},
			&cli.StringFlag{
				Name:    "file, f",
				Aliases: []string{"f"},
				Usage:   "path of Dockerfile or app.yaml",
			},
			&cli.StringFlag{
				Name:    "domain",
				Aliases: []string{"d"},
				Usage:   "domain name to use with this app",
			},
		},
	})
}

func cmdDeploy(c *cli.Context) error {
	dir, name, err := getDirApp(".")
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	app, err := client.GetApp(name)
	if err != nil {
		log.Warn(err)
		return app404Err(name)
	}

	var build *pb.Build
	buildId := c.String("build")

	if len(buildId) == 0 {
		build, err = executeBuildDir(client, app, dir)
		stdcli.ExitOnError(err)
	} else {
		b, err := client.GetBuild(name, buildId)
		stdcli.ExitOnError(err)

		if b == nil {
			err = fmt.Errorf("No build found by id: %s.", buildId)
			stdcli.ExitOnError(err)
		}

		build = b
	}

	fmt.Printf("Deploying build %s\n", build.Id)

	_, err = client.ReleaseBuild(build, pb.ReleaseOptions{
		Domain: c.String("domain"),
		Wait:   c.Bool("wait"),
	})
	stdcli.ExitOnError(err)

	app, err = client.GetApp(name)
	if err != nil {
		err = fmt.Errorf("fetching app %s err: %v", name, err)
		stdcli.ExitOnError(err)
	}

	if len(app.Endpoint) > 0 {
		fmt.Printf("\nDeployed at %s\n", app.Endpoint)
	} else {
		fmt.Println("[DONE].")
	}

	return nil
}
