package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/crypto/rand"
	term "github.com/appscode/go/term"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cmd/provider/aws"
	"github.com/dinesh/datacol/cmd/provider/gcp"
	"github.com/dinesh/datacol/cmd/stdcli"
	"github.com/dinesh/datacol/go/env"
	"github.com/urfave/cli"
)

var (
	credNotFound    = errors.New("Invalid credentials")
	projectNotFound = errors.New("Invalid project id")

	defaultGcpZone         = "asia-east1-a" //Taiwan
	defaultAWSZone         = "ap-southeast-1a"
	defaultAWSInstanceType = "t2.medium"
	defaultGCPInstanceType = "n1-standard-1"
)

const (
	serviceAccountKey = "service_account"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:        "init",
		Usage:       "[cloud-provider] [credentials.csv]",
		Description: "create new datacol stack",
		Action:      cmdStackCreate,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Name of stack",
				Value: "demo",
			},
			&cli.StringFlag{
				Name:  "region",
				Usage: "region for stack",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "zone",
				Usage: "default availability zone for stack",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "bucket",
				Usage: "storage bucket",
			},
			&cli.IntFlag{
				Name:  "nodes",
				Usage: "number of nodes in container cluster",
				Value: 2,
			},
			&cli.StringFlag{
				Name:  "cluster",
				Usage: "name for existing Kubernetes cluster (if present)",
			},
			&cli.IntFlag{
				Name:  "disk-size",
				Usage: "SSD disk size for cluster in GB",
				Value: 10,
			},
			&cli.StringFlag{
				Name:  "machine-type",
				Usage: "type of instance to use for cluster nodes",
				Value: "",
			},
			&cli.BoolFlag{
				Name:  "preemptible",
				Usage: "use preemptible vm (for GCP)",
			},
			&cli.BoolFlag{
				Name:  "opt-out",
				Usage: "Opt-out from getting updates via email from `datacol`",
			},
			&cli.StringFlag{
				Name:  "password",
				Usage: "api password for the stack",
			},
			&cli.StringFlag{
				Name:  "cluster-version",
				Usage: "The Kubernetes version to use for the master and nodes",
				Value: "1.7.11-gke.1",
			},
			&cli.StringFlag{
				Name:  "key",
				Usage: "[Name of ssh-keypair to create for AWS] OR [/path/to/service-account key for GCP]",
			},
		},
	})

	stdcli.AddCommand(cli.Command{
		Name:   "destroy",
		Usage:  "destroy the datacol stack from your cloud account",
		Action: cmdStackDestroy,
	})
}

func cmdStackCreate(c *cli.Context) (err error) {
	if c.NArg() < 1 {
		err = fmt.Errorf("Please provide a cloud provider (aws, gcp, local)")
	}

	stdcli.ExitOnError(err)

	provider := c.Args().Get(0)
	switch strings.ToLower(provider) {
	case "gcp":
		err = cmdGCPStackCreate(c)
	case "aws":
		err = cmdAWSStackCreate(c)
	case "local":
		err = cmdLocalStackCreate(c)
	default:
		err = fmt.Errorf("Invalid cloud provider: %s. Should be either of `aws` or `gcp`.", provider)
	}

	stdcli.ExitOnError(err)
	return
}

func cmdAWSStackCreate(c *cli.Context) error {
	var credentialsFile string
	if c.NArg() > 1 {
		credentialsFile = c.Args().Get(1)
	}

	stackName := c.String("name")
	options := &aws.InitOptions{
		Name:            stackName,
		DiskSize:        c.Int("disk-size"),
		NumNodes:        c.Int("nodes"),
		MachineType:     c.String("machine-type"),
		Zone:            c.String("zone"),
		Region:          c.String("region"),
		Bucket:          c.String("bucket"),
		Version:         stdcli.Version,
		ApiKey:          c.String("ApiKey"),
		KeyName:         c.String("key"),
		UseSpotInstance: c.Bool("preemptible"),
		CreateCluster:   len(c.String("cluster")) == 0,
	}

	if len(options.ApiKey) == 0 {
		options.ApiKey = rand.GeneratePassword()
	}

	ec := env.FromHost()
	if ec.DevMode() {
		options.ArtifactBucket = "datacol-dev"
	} else {
		options.ArtifactBucket = "datacol-distros"
	}

	if err := initializeAWS(options, credentialsFile); err != nil {
		return err
	}

	term.Successln("\nDONE")
	fmt.Printf("Next, create an app with `STACK=%s datacol apps create`.\n", stackName)
	return nil
}

func cmdLocalStackCreate(c *cli.Context) error {
	auth := &stdcli.Auth{
		Name:      c.String("name"),
		ApiServer: "localhost",
		Provider:  "local",
	}

	return stdcli.SetAuth(auth)
}

func cmdGCPStackCreate(c *cli.Context) error {
	stackName := c.String("name")
	zone := c.String("zone")
	nodes := c.Int("nodes")
	bucket := c.String("bucket")
	password := c.String("password")

	cluster := c.String("cluster")
	machineType := c.String("machine-type")
	preemptible := c.Bool("preemptible")
	diskSize := c.Int("disk-size")

	options := &gcp.InitOptions{
		Name:           stackName,
		ClusterName:    cluster,
		DiskSize:       diskSize,
		NumNodes:       nodes,
		MachineType:    machineType,
		Zone:           zone,
		Bucket:         bucket,
		Preemptible:    preemptible,
		Version:        stdcli.Version,
		ApiKey:         password,
		SAKeyPath:      c.String("key"),
		ClusterVersion: c.String("cluster-version"),
	}

	ec := env.FromHost()

	if options.Zone == "" {
		options.Zone = defaultGcpZone
	}

	if options.Region == "" {
		options.Region = getGcpRegionFromZone(options.Zone)
	}

	if options.MachineType == "" {
		if ec.DevMode() {
			options.MachineType = "g1-small"
		} else {
			options.MachineType = defaultGCPInstanceType
		}
	}

	if len(options.Bucket) == 0 {
		options.Bucket = fmt.Sprintf("datacol-%s", options.Name)
	}

	if len(options.ApiKey) == 0 {
		options.ApiKey = rand.GeneratePassword()
	}

	if ec.DevMode() {
		options.ArtifactBucket = "datacol-dev"
		options.Preemptible = true
	} else {
		options.ArtifactBucket = "datacol-distros"
	}

	if err := initializeGCP(options, nodes, c.Bool("opt-out")); err != nil {
		stdcli.ExitOnError(err)
	}

	term.Successln("\nDONE")

	fmt.Printf("Next, create an app with `STACK=%s datacol apps create`.\n", stackName)
	return nil
}

func initializeAWS(opts *aws.InitOptions, credentialsFile string) error {
	if opts.Zone == "" {
		opts.Zone = defaultAWSZone
	}

	if opts.Region == "" {
		opts.Region = getAwsRegionFromZone(opts.Zone)
	}

	if opts.MachineType == "" {
		opts.MachineType = defaultAWSInstanceType
	}

	if len(opts.Bucket) == 0 {
		opts.Bucket = fmt.Sprintf("datacol-%s-settings", opts.Name)
	}

	creds, err := aws.ReadCredentials(credentialsFile)
	if err != nil {
		return err
	}

	if creds == nil {
		return err
	}

	fmt.Println("Using AWS Access Key ID:", creds.Access)

	if err := saveAwsCredential(opts.Name, creds); err != nil {
		return err
	}

	ret, err := aws.InitializeStack(opts, creds)
	if err != nil {
		return err
	}

	if err = saveKeyPairData(opts.Name, ret.KeyPairData); err != nil {
		return err
	}

	if err = dumpAwsAuthParams(opts.Name, opts.Region, opts.Bucket, ret.Host, ret.Password); err != nil {
		return err
	}

	fmt.Printf("\nStack hostIP %s\n", ret.Host)
	fmt.Printf("Stack password: %s [Please keep is secret]\n", ret.Password)
	fmt.Println("The above configuration has been saved in your home directory at ~/.datacol/config.json")

	return nil
}

func initializeGCP(opts *gcp.InitOptions, nodes int, optout bool) error {
	var cred []byte

	if opts.SAKeyPath != "" {
		data, err := ioutil.ReadFile(opts.SAKeyPath)
		if err != nil {
			return err
		}
		cred = data

		jwtConfig, err := parseGCPCredentialsFile(data)
		if err != nil {
			return err
		}

		opts.Project = jwtConfig.ProjectID
		opts.SAEmail = jwtConfig.ClientEmail
	} else {
		resp := gcp.CreateCredential(opts.Name, optout)
		if resp.Err != nil {
			return resp.Err
		}

		cred = resp.Cred
		if len(cred) == 0 {
			return credNotFound
		}

		if len(resp.ProjectId) == 0 {
			return projectNotFound
		}

		opts.Project = resp.ProjectId
		opts.SAEmail = resp.SAEmail
	}

	if err := saveGcpCredential(opts.Name, cred); err != nil {
		return err
	}

	if len(opts.Bucket) == 0 {
		opts.Bucket = fmt.Sprintf("datacol-%s", slug(opts.Project))
	}

	name := opts.Name
	if len(opts.ClusterName) == 0 {
		opts.ClusterNotExists = true
		opts.ClusterName = fmt.Sprintf("%v-cluster", name)
	} else {
		opts.ClusterNotExists = false
	}

	apis := []string{
		"datastore.googleapis.com",
		"cloudbuild.googleapis.com",
		"deploymentmanager",
		"iam.googleapis.com",
	}

	url := fmt.Sprintf("https://console.cloud.google.com/flows/enableapi?apiid=%s&project=%s", strings.Join(apis, ","), opts.Project)

	fmt.Printf("\nDatacol needs to communicate with various APIs provided by cloud platform, please enable APIs by opening following link in browser and click Continue: \n%s\n", url)
	term.Confirm("Are you done ?")

	res, err := gcp.InitializeStack(opts)
	if err != nil {
		return err
	}

	fmt.Printf("\nStack hostIP %s\n", res.Host)
	fmt.Printf("Stack password: %s [Please keep is secret]\n", res.Password)
	fmt.Println("The above configuration has been saved in your home directory at ~/.datacol/config.json")

	return dumpGcpAuthParams(opts.Name, opts.Project, opts.Bucket, res.Host, res.Password)
}

func cmdStackDestroy(c *cli.Context) (err error) {
	auth, _ := stdcli.GetAuthOrDie()
	provider := auth.Provider

	prompt := fmt.Sprintf("This is destructive action. Do you want to delete %s stack on %s ?", auth.Name, provider)
	if !term.Ask(prompt, false) {
		return nil
	}

	switch strings.ToLower(provider) {
	case "gcp":
		err = gcpTeardown()
	case "aws":
		err = awsTeardown()
	default:
		err = fmt.Errorf("Invalid cloud provider: %s. Should be either of aws or gcp.", provider)
	}

	stdcli.ExitOnError(err)
	term.Successln("\nDONE")
	return nil
}

func awsTeardown() error {
	auth, rc := stdcli.GetAuthOrDie()
	var credentialsFile string

	credentialsFile = filepath.Join(pb.ConfigPath, auth.Name, pb.AwsCredentialFile)
	creds, err := aws.ReadCredentials(credentialsFile)

	if err != nil {
		return err
	}

	if creds == nil {
		return err
	}

	if err := aws.TeardownStack(auth.Name, auth.Bucket, auth.Region, creds); err != nil {
		return err
	}

	if err := rc.DeleteAuth(); err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(pb.ConfigPath, auth.Name))
}

func gcpTeardown() error {
	auth, rc := stdcli.GetAuthOrDie()
	if err := gcp.TeardownStack(auth.Name, auth.Project, auth.Bucket); err != nil {
		return err
	}

	if err := rc.DeleteAuth(); err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(pb.ConfigPath, auth.Name))
}

func createStackDir(name string) error {
	cfgroot := filepath.Join(pb.ConfigPath, name)
	return os.MkdirAll(cfgroot, 0700)
}

func saveGcpCredential(name string, data []byte) error {
	if err := createStackDir(name); err != nil {
		return err
	}
	path := filepath.Join(pb.ConfigPath, name, pb.SvaFilename)
	log.Debugf("saving GCP credentials at %s", path)

	return ioutil.WriteFile(path, data, 0700)
}

func saveKeyPairData(name, content string) error {
	path := filepath.Join(pb.ConfigPath, name, pb.AwsKeyPemPath)
	log.Debugf("saving keypair at %s", path)
	return ioutil.WriteFile(path, []byte(content), 0700)
}

func saveAwsCredential(name string, cred *aws.AwsCredentials) error {
	if err := createStackDir(name); err != nil {
		return err
	}
	path := filepath.Join(pb.ConfigPath, name, pb.AwsCredentialFile)
	log.Debugf("saving AWS credentials at %s", path)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	wr := csv.NewWriter(file)
	defer wr.Flush()

	wr.Write([]string{"AWSAccessKeyId", "AWSSecretKey"})
	wr.Write([]string{cred.Access, cred.Secret})

	if err = wr.Error(); err != nil {
		return fmt.Errorf("writing csv err: %v", err)
	}
	return nil
}

func dumpAwsAuthParams(name, region, bucket, host, api_key string) error {
	auth := &stdcli.Auth{
		Name:      name,
		ApiServer: host,
		Region:    region,
		ApiKey:    api_key,
		Bucket:    bucket,
		Provider:  "aws",
	}

	return stdcli.SetAuth(auth)
}

func dumpGcpAuthParams(name, project, bucket, host, api_key string) error {
	auth := &stdcli.Auth{
		Name:      name,
		Project:   project,
		Bucket:    bucket,
		ApiServer: host,
		ApiKey:    api_key,
		Provider:  "gcp",
	}

	return stdcli.SetAuth(auth)
}

func getAwsRegionFromZone(zone string) string {
	return zone[0 : len(zone)-1]
}

func getGcpRegionFromZone(zone string) string {
	return zone[0 : len(zone)-2]
}

type credentialsFile struct {
	Type string `json:"type"` // serviceAccountKey or userCredentialsKey

	// Service Account fields
	ClientEmail string `json:"client_email"`
	ProjectID   string `json:"project_id"`
}

func parseGCPCredentialsFile(jsonKey []byte) (*credentialsFile, error) {
	var f credentialsFile
	if err := json.Unmarshal(jsonKey, &f); err != nil {
		return nil, err
	}
	if f.Type != "service_account" {
		return nil, fmt.Errorf("google: read JWT from JSON credentials: 'type' field is %q (expected %q)", f.Type, serviceAccountKey)
	}

	return &f, nil
}
