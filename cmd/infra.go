package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/dinesh/datacol/client"
	"github.com/dinesh/datacol/cmd/stdcli"
	"github.com/urfave/cli"

	log "github.com/Sirupsen/logrus"
)

type ResourceType struct {
	name, gcpArgs, awsArgs string
}

var resourceTypes = []ResourceType{
	{
		name:    "mysql",
		gcpArgs: "--tier=db-g1-small,--activation-policy=ALWAYS,--db-version=MYSQL_5_7",
		awsArgs: "--allocated-storage=10,--database=app,--instance-type=db.t2.micro,--multi-az=false,--password,--private=false,--username=app,--version=5.7.16",
	},
	{
		name:    "postgres",
		gcpArgs: "--cpu=1,--memory=4096,--db-version=POSTGRES_9_6,--activation-policy=ALWAYS",
		awsArgs: "--allocated-storage=10,--database=app,--instance-type=db.t2.micro,--max-connections={DBInstanceClassMemory/15000000},--multi-az=false,--password=,--private,--username=app,--version=9.6.1",
	},
	{
		name:    "redis",
		gcpArgs: "",
		awsArgs: "--automatic-failover-enabled,--database=app,--instance-type=cache.t2.micro,--num-cache-clusters=1,--private=false",
	},
}

func init() {
	stdcli.AddCommand(cli.Command{
		Name:   "infra",
		Usage:  "Managed GCP stack resources and infrastructure",
		Action: cmdResourceList,
		Subcommands: []cli.Command{
			{
				Name:   "create",
				Usage:  "create a new resource",
				Action: cmdResourceCreate,
				Flags: []cli.Flag{
					stackFlag,
					&cli.BoolFlag{Name: "wait, w"},
				},
				SkipFlagParsing: true,
			},
			{
				Name:      "delete",
				ArgsUsage: "[name]",
				Usage:     "delete a existing resource",
				Action:    cmdResourceDelete,
				Flags:     []cli.Flag{stackFlag},
			},
			{
				Name:      "info",
				ArgsUsage: "[name]",
				Usage:     "get info for a existing resource",
				Action:    cmdResourceInfo,
				Flags:     []cli.Flag{stackFlag},
			},
			{
				Name:   "link",
				Usage:  "link an app to resource and setting it up.",
				Action: cmdLinkCreate,
				Flags:  []cli.Flag{appFlag},
			},
			{
				Name:   "unlink",
				Usage:  "unlink an resource from an app.",
				Action: cmdLinkDelete,
				Flags:  []cli.Flag{appFlag},
			},
		},
	})
}

func cmdResourceList(c *cli.Context) error {
	api, close := getApiClient(c)
	defer close()

	resp, err := api.ListResources()
	if err != nil {
		return err
	}

	fmt.Printf("Cloud: %s\n", api.Provider)
	fmt.Printf("Name: %s\n", api.Name)

	fmt.Printf("Datacol API Host: %s\n", api.ApiServer)
	fmt.Printf("Datacol API Key: %s\n", api.ApiKey)

	if api.IsGCP() {
		fmt.Printf("GCP Project: %s\n", api.Project)
		fmt.Printf("GCP Region: %s\n", api.Region)
	}

	if api.IsAWS() {
		fmt.Printf("AWS Region: %s\n", api.Region)
	}

	fmt.Println("\nResource:")

	for _, r := range resp {
		kind := r.Kind
		if _, err := checkResourceType(kind); err == nil {
			fmt.Println(r.Name)
		}
	}

	return nil
}

func cmdResourceInfo(c *cli.Context) error {
	name := c.Args().Get(0)
	if len(name) == 0 {
		log.Errorf("no name given. Usage: \n")
		stdcli.Usage(c)
	}

	api, close := getApiClient(c)
	defer close()

	rs, err := api.GetResource(name)
	stdcli.ExitOnError(err)

	// for k, v := range jsonDecode(rs.Exports) {
	// 	fmt.Printf("%s=%s", k, v)
	// }

	fmt.Printf("%s", toJson(rs))

	fmt.Printf("\n")
	return nil
}

func cmdResourceDelete(c *cli.Context) error {
	name := c.Args().Get(0)
	if len(name) == 0 {
		log.Errorf("no name given. Usage: \n")
		stdcli.Usage(c)
	}
	api, close := getApiClient(c)
	defer close()

	stdcli.ExitOnError(api.DeleteResource(name))

	if api.IsAWS() {
		stdcli.ExitOnError(waitForAwsResource(name, "DELETE", api))
	}

	if api.IsGCP() {
		stdcli.ExitOnError(waitForGcpResource(name, "DELETE", api))
	}

	fmt.Println("\nDELETED")
	return nil
}

func cmdResourceCreate(c *cli.Context) error {
	api, close := getApiClient(c)
	defer close()

	t, err := checkResourceType(c.Args().First())
	stdcli.ExitOnError(err)

	var arguments string
	if api.IsGCP() {
		arguments = t.gcpArgs
	} else {
		arguments = t.awsArgs
	}

	args := append(strings.Split(arguments, ","), c.Args().Tail()...)
	stdcli.EnsureOnlyFlags(c, args)
	options := stdcli.FlagsToOptions(c, args)

	var optionsList []string
	for key, val := range options {
		optionsList = append(optionsList, fmt.Sprintf("%s=%q", key, val))
	}

	if options["name"] == "" {
		options["name"] = withIntSuffix(t.name)
	}

	if v, ok := options["password"]; ok && len(v) <= 8 {
		options["password"] = rand.GeneratePassword()
	}

	fmt.Printf("Creating %s (%s", options["name"], t.name)
	if len(optionsList) > 0 {
		sort.Strings(optionsList)
		fmt.Printf(": %s", strings.Join(optionsList, " "))
	}
	fmt.Printf(")... ")
	fmt.Printf("\n")

	rs, err := api.CreateResource(t.name, options)
	stdcli.ExitOnError(err)

	log.Debugf("Resource: %v", toJson(rs))

	if api.IsAWS() {
		stdcli.ExitOnError(waitForAwsResource(options["name"], "CREATE", api))
	}

	if api.IsGCP() {
		stdcli.ExitOnError(waitForGcpResource(options["name"], "CREATE", api))
	}

	fmt.Println("\nCREATED")
	return nil
}

func cmdLinkCreate(c *cli.Context) error {
	_, app, err := getDirApp(".")
	if err != nil {
		return err
	}
	name := c.Args().First()

	client, close := getApiClient(c)
	defer close()

	stdcli.ExitOnError(client.CreateResourceLink(app, name))

	fmt.Printf("Linked %s to %s\n", name, app)
	return nil
}

func cmdLinkDelete(c *cli.Context) error {
	_, app, err := getDirApp(".")
	stdcli.ExitOnError(err)

	name := c.Args().First()
	client, close := getApiClient(c)

	defer close()

	stdcli.ExitOnError(client.DeleteResourceLink(app, name))

	fmt.Printf("Deleted link %s from %s\n", name, app)
	return nil
}

func checkResourceType(t string) (*ResourceType, error) {
	for _, resourceType := range resourceTypes {
		if resourceType.name == t {
			return &resourceType, nil
		}
	}

	return nil, fmt.Errorf("unsupported resource type %s; see 'datacol infra create --help'", t)
}

func jsonDecode(b []byte) map[string]string {
	var opts map[string]string
	if err := json.Unmarshal(b, &opts); err != nil {
		log.Fatal(fmt.Errorf("unmarshaling %+v err:%v", opts, err))
	}
	return opts
}

func waitForAwsResource(name, event string, c *client.Client) error {
	tick := time.Tick(time.Second * 2)
	timeout := time.After(time.Minute * 5)
	fmt.Printf("Waiting for %s ", name)
	failedEv := event + "_FAILED"
	completedEv := event + "_COMPLETE"
Loop:
	for {
		select {
		case <-tick:
			rs, err := c.GetResource(name)
			if err != nil {
				if event == "DELETE" {
					return nil
				}
				return err
			}

			fmt.Print(".")
			if rs.Status == failedEv {
				return fmt.Errorf("%s failed because of \"%s\"", event, rs.StatusReason)
			}
			if rs.Status == completedEv {
				break Loop
			}
		case <-timeout:
			fmt.Print("timeout (5 minutes). Skipping")
			break Loop
		}
	}

	return nil
}

func waitForGcpResource(name, event string, c *client.Client) error {
	tick := time.Tick(time.Second * 2)
	timeout := time.After(time.Minute * 5)
	fmt.Printf("Waiting for %s ", name)

	/*
		NOTE: Though GCP deployment manager provides status=PENDING|RUNNING|DONE, we add an additional "FAILED" value to denote
		failed event which is equivalent to DONE with Operation.Error != nil.

		This is set in cloud/gcp/resource.go#resourceFromDeployment.
	*/
	failedEvent := "FAILED"

Loop:
	for {
		select {
		case <-tick:
			rs, err := c.GetResource(name)
			if err != nil {
				if event == "DELETE" {
					return nil
				}
				return err
			}

			fmt.Print(".")
			if rs.Status == failedEvent {
				return fmt.Errorf("%s failed because of \"%s\"", event, rs.StatusReason)
			}

			if rs.Status == "DONE" {
				break Loop
			}
		case <-timeout:
			fmt.Print("timeout (5 minutes). Skipping")
			break Loop
		}
	}

	return nil
}
