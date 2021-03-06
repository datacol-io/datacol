package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go/term"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "deploy",
		Usage:  "deploy an app",
		Action: cmdDeploy,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "image, i",
				Usage: "docker image tag to use",
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
			&cli.StringFlag{
				Name:  "ref",
				Usage: "The commit SHA1 of branch or tag to use",
			},
			&cli.BoolTFlag{
				Name:  "wait, w",
				Usage: "Wait for the app become available",
			},
			&cli.StringFlag{
				Name:  "file, f",
				Usage: "path of Dockerfile or app.yaml",
			},
			&cli.BoolTFlag{
				//TODO: support expose in API
				Name:  "expose",
				Usage: "expose the service to the public",
			},
			&appFlag,
		},
	})
}

func cmdDeploy(c *cli.Context) error {
	dir, name, err := getDirApp(".", c)
	stdcli.ExitOnError(err)

	client, close := getApiClient(c)
	defer close()

	app, err := client.GetApp(name)
	if err != nil {
		log.Warn(err)
		return app404Err(name)
	}

	var build *pb.Build
	commitID, buildID, image := c.String("ref"), c.String("build"), c.String("image")

	if len(buildID) == 0 {
		if image != "" {
			build, err = executeBuildImage(client, app, image, commitID)
		} else {
			r, err := stdinInput(c)
			if err != nil {
				return err
			}

			if r != nil {
				build, err = executeBuildDockerArchive(client, app, r, commitID)
			} else if commitID == "" {
				build, err = executeBuildDir(client, app, dir, buildID)
			} else {
				build, err = executeBuildGitSource(client, app, commitID, buildID)
			}
		}

		stdcli.ExitOnError(err)
	} else {
		b, err := client.GetBuild(name, buildID)
		stdcli.ExitOnError(err)

		if b == nil {
			err = fmt.Errorf("No build found by id: %s.", buildID)
			stdcli.ExitOnError(err)
		}

		build = b
	}

	build, err = client.GetBuild(name, build.Id)
	stdcli.ExitOnError(err)

	if build.Status == "FAILED" {
		term.Fatalln(fmt.Sprintf("BUILD=%s is having FAILED status.", buildID))
	}

	fmt.Printf("Deploying build %s\n", build.Id)

	_, err = client.ReleaseBuild(build, pb.ReleaseOptions{
		Wait:   c.Bool("wait"),
		Expose: c.BoolT("expose"),
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
