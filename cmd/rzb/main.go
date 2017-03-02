package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dinesh/rz/client/build"
	// "github.com/dinesh/rz/client/deploy"

	"github.com/dinesh/rz/client"
	"gopkg.in/urfave/cli.v2"
)

func checkFlags(c *cli.Context, flags ...string) error {
	for _, flag := range flags {
		value := c.String(flag)
		if value == "" {
			return fmt.Errorf("missing required option: --%v", flag)
		}
	}

	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	doBuild := func(c *cli.Context) error {
		var appDir string

		if c.NArg() > 0 {
			appDir = c.Args().Get(0)
		} else {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			appDir = dir
		}

		if appDir == "" {
			return errors.New("missing project directory")
		}

		auth, app, err := client.SetApp(appDir)
		if err != nil {
			return err
		}

		return build.ExecuteBuildDir(app, auth)
	}

	doBuildCancel := func(c *cli.Context) error {
		var buildId string
		if c.NArg() > 0 {
			buildId = c.Args().Get(0)
		} else {
			return errors.New("no buildId gived")
		}

		_, auth := client.GetAuthOrDie()
		return build.CancelBuild(auth, buildId)
	}

	doAppDeploy := func(c *cli.Context) error {
		_, auth := client.GetAuthOrDie()

		// _, err := deploy.NewDeployer("", "", true)
		// if err != nil {
		// 	return err
		// }

		fmt.Printf("current Auth: %s", auth.RackName)
		return nil
	}

	doStackCreate := func(c *cli.Context) error {
		if err := checkFlags(c, "zone", "bucket"); err != nil {
			return err
		}

		_, auth := client.GetAuthOrDie()
		auth.Zone = c.String("zone")
		auth.BucketName = c.String("bucket")

		if err := client.SetAuth(auth); err != nil {
			return err
		}

		numNodes := c.Int("nodes")
		if numNodes == 0 {
			numNodes = 3
		}

		return client.CreateStack(auth, numNodes)
	}

	doStackDelete := func(c *cli.Context) error {
		_, auth := client.GetAuthOrDie()
		return client.DeleteStack(auth)
	}

	doAppDelete := func(c *cli.Context) error {
		rc, auth := client.GetAuthOrDie()
		auth.DeleteApp(c.String("app"))
		
		if err := rc.SetAuth(auth); err != nil {
			return err
		}
		return rc.Write()
	}

	doLogin := func(c *cli.Context) error {
		rack := c.String("rack")
		projectId := c.String("project-id")

		if rack == "" {
			rack = "razorbox-demo"
		}

		client.CreateGCECredential(rack, projectId)
		return nil
	}

	app := &cli.App{
		Commands: []cli.Command{
			cli.Command{
				Name:  "stack",
				UsageText: "manage stacks on GCP",
				Subcommands: []cli.Command{
					{
						Name: "create",
						UsageText: "create a new stack",
						Action: doStackCreate,
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "zone"},
							&cli.StringFlag{Name: "bucket"},
							&cli.IntFlag{Name: "nodes"},
						},
					},
					{
						Name: "delete",
						UsageText: "delete a stack",
						Action: doStackDelete,
					},
				},
			},
			cli.Command{
				Name:      "build",
				UsageText: "build an app",
				Action:    doBuild,
				Subcommands:  []cli.Command {
					{
						Name:  "cancel",
						Usage: "cancel an ongoing build",
						Action: doBuildCancel,
					},
				},
			},
			cli.Command{
				Name:      "delete",
				UsageText: "delete an app",
				Action:    doAppDelete,
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "app, a"},
				},
			},
			cli.Command{
				Name:      "deploy",
				UsageText: "deploy an app",
				Action:    doAppDeploy,
				Flags:     []cli.Flag{},
			},
			cli.Command{
				Name:      "login",
				UsageText: "login with razorbox",
				Action:    doLogin,
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "rack"},
					&cli.StringFlag{Name: "project-id"},
				},
			},
		},
	}

	app.Run(os.Args)
}
