package main

import (
	"fmt"
	"github.com/dinesh/datacol/client"
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "init",
		Usage:  "create new stack",
		Action: cmdStackCreate,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "stack",
				Usage: "Name of stack",
				Value: "dev",
			},
			&cli.StringFlag{
				Name:  "project",
				Usage: "GCP project name or id to use",
			},
			&cli.StringFlag{
				Name:  "zone",
				Usage: "zone for stack",
				Value: "us-east1-b",
			},
			&cli.StringFlag{
				Name:  "bucket",
				Usage: "GCP storage bucket",
			},
			&cli.IntFlag{
				Name:  "nodes",
				Usage: "number of nodes in container cluster",
				Value: 2,
			},
			&cli.StringFlag{
				Name:  "cluster",
				Usage: "name for existing Kubernetes cluster in GCP",
			},
			&cli.StringFlag{
				Name:  "machine-type",
				Usage: "name of machine-type to use for cluster",
				Value: "n1-standard-1",
			},
			&cli.BoolTFlag{
				Name:  "preemptible",
				Usage: "use preemptible vm",
			},
			&cli.BoolFlag{
				Name:  "opt-out",
				Usage: "Opt-out from getting updates by email by `datacol`",
			},
		},
	})

	stdcli.AddCommand(cli.Command{
		Name:      "destroy",
		UsageText: "destroy current stack",
		Action:    cmdStackDestroy,
	})
}

func cmdStackCreate(c *cli.Context) error {
	stdcli.CheckFlagsPresence(c, "project")

	stackName := c.String("stack")
	project := c.String("project")
	zone := c.String("zone")
	nodes := c.Int("nodes")
	bucket := c.String("bucket")

	if len(bucket) == 0 {
		bucket = fmt.Sprintf("datacol-%s", slug(project))
	}

	cluster := c.String("cluster")
	machineType := c.String("machine-type")
	preemptible := c.Bool("preemptible")

	ac := getAnonClient(c)
	st, err := client.FindStack(stackName)

	if err != nil {
		//todo: handler err better, 1. formatting error 2) no stack found
		ac.StackName = stackName

		if st, err = ac.CreateStack(project, zone, bucket, c.BoolT("opt-out")); err != nil {
			return err
		}
	}

	ac.SetStack(st.Name)

	fmt.Printf("Please ENABLE following apis in your Google Account: \n")
	confirm(fmt.Sprintf("[Deployment Manager V2]: Open follwoing link in browser\n %s", consoleURL("deploymentmanager", st.ProjectId)), 2)
	confirm(fmt.Sprintf("\n[Container Builder]: Open follwoing link in browser\n %s", consoleURL("cloudbuild.googleapis.com", st.ProjectId)), 2)

	// sleep for IAM permissons to resolve before getting token
	// time.Sleep(time.Second * 2)

	if err = ac.DeployStack(st, cluster, machineType, nodes, preemptible); err != nil {
		return err
	}

	fmt.Printf("\nDONE\n")
	return nil
}

func cmdStackDestroy(c *cli.Context) error {
	client := getClient(c)
	if err := client.DestroyStack(); err != nil {
		return err
	}
	fmt.Printf("\nDONE\n")
	return nil
}
