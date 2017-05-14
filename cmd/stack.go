package main

import (
	"errors"
	"fmt"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cmd/provider/gcp"
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	credNotFound = errors.New("Invalid credentials")
)

func init() {
	stdcli.AddCommand(&cli.Command{
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
				Usage: "GCP project id to use",
			},
			&cli.StringFlag{
				Name:  "zone",
				Usage: "GCP zone for stack",
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
			&cli.BoolFlag{
				Name:  "preemptible",
				Usage: "use preemptible vm",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "opt-out",
				Usage: "Opt-out from getting updates by email by `datacol`",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "api password for the stack",
			},
		},
	})

	stdcli.AddCommand(&cli.Command{
		Name:   "destroy",
		Usage:  "destroy a stack from GCP",
		Action: cmdStackDestroy,
	})
}

func cmdStackCreate(c *cli.Context) error {
	stdcli.CheckFlagsPresence(c, "project")

	stackName := c.String("stack")
	project := c.String("project")
	zone := c.String("zone")
	nodes := c.Int("nodes")
	bucket := c.String("bucket")
	password := c.String("password")

	if len(bucket) == 0 {
		bucket = fmt.Sprintf("datacol-%s", slug(project))
	}

	cluster := c.String("cluster")
	machineType := c.String("machine-type")
	preemptible := c.Bool("preemptible")

	message := `Welcome to Datacol CLI. This command will guide you through creating a new infrastructure inside your Google account. 
It uses various Google services (like Container engine, Cloudbuilder, Deployment Manager etc) under the hood to 
automate all away to give you a better deployment experience.

Datacol CLI will authenticate with your Google Account and install the Datacol platform into your GCP account. 
These credentials will only be used to communicate between this installer running on your computer and the Google platform.
`

	fmt.Printf(message)

	apis := []string{
		"datastore.googleapis.com",
		"cloudbuild.googleapis.com",
		"deploymentmanager",
		"iam.googleapis.com",
	}

	fmt.Printf("\nDatacol needs to communicate with various APIs provided by cloud platform, please enable APIs by opening following link in browser and click Continue.\n")
	url := fmt.Sprintf("https://console.cloud.google.com/flows/enableapi?apiid=%s&project=%s", strings.Join(apis, ","), project)
	prompt(url)

	options := &gcp.InitOptions{
		Name:        stackName,
		ClusterName: cluster,
		DiskSize:    10,
		NumNodes:    nodes,
		MachineType: machineType,
		Zone:        zone,
		Bucket:      bucket,
		Preemptible: preemptible,
		Project:     project,
		Version:     stdcli.Version,
		API_KEY:     password,
	}

	if err := initialize(options, nodes, c.Bool("opt-out")); err != nil {
		return err
	}

	fmt.Printf("\nDONE.\n")

	fmt.Printf("Next, create an app with `STACK=%s datacol apps create`.\n", stackName)
	return nil
}

func cmdStackDestroy(c *cli.Context) error {
	if err := teardown(); err != nil {
		return err
	}

	fmt.Printf("\nDONE\n")
	return nil
}

func initialize(opts *gcp.InitOptions, nodes int, optout bool) error {
	resp := gcp.CreateCredential(opts.Name, opts.Project, optout)
	if resp.Err != nil {
		return resp.Err
	}

	cred := resp.Cred
	if len(cred) == 0 {
		return credNotFound
	}

	if err := saveCredential(opts.Name, cred); err != nil {
		return err
	}

	opts.Project = resp.ProjectId
	opts.ProjectNumber = resp.PNumber
	opts.SAEmail = resp.SAEmail

	name := opts.Name
	if len(opts.ClusterName) == 0 {
		opts.ClusterNotExists = true
		opts.ClusterName = fmt.Sprintf("%v-cluster", name)
	} else {
		opts.ClusterNotExists = false
	}

	time.Sleep(2 * time.Second) // wait for sometime for iam permission propagation

	res, err := gcp.InitializeStack(opts)
	if err != nil {
		return err
	}

	fmt.Printf("\nStack hostIP %s\n", res.Host)
	fmt.Printf("Stack password: %s [Please keep is secret]\n", res.Password)

	return dumpParams(opts.Name, opts.Project, opts.Bucket, res.Host, res.Password)
}

func teardown() error {
	name := stdcli.CurrentStack()
	project := stdcli.ReadSetting(name, "project")
	bucket := stdcli.ReadSetting(name, "bucket")

	if err := gcp.TeardownStack(name, project, bucket); err != nil {
		return err
	}

	os.Remove(filepath.Join(pb.ConfigPath, "stack"))
	return os.RemoveAll(filepath.Join(pb.ConfigPath, name))
}

func createStackDir(name string) error {
	cfgroot := filepath.Join(pb.ConfigPath, name)
	if err := os.MkdirAll(cfgroot, 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(pb.ConfigPath, "stack"), []byte(name), 0700); err != nil {
		return err
	}
	return nil
}

func saveCredential(name string, data []byte) error {
	if err := createStackDir(name); err != nil {
		return err
	}

	path := filepath.Join(pb.ConfigPath, name, pb.SvaFilename)
	return ioutil.WriteFile(path, data, 0777)
}

func dumpParams(name, project, bucket, host, api_key string) error {
	stdcli.WriteSetting(name, "project", project)
	stdcli.WriteSetting(name, "api_key", api_key)
	stdcli.WriteSetting(name, "api_host", host)
	return stdcli.WriteSetting(name, "bucket", bucket)
}
