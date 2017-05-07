package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/dinesh/datacol/cmd/stdcli"
	"gopkg.in/urfave/cli.v2"
	"math/rand"
	"sort"
	"strings"
	"time"
)

type ResourceType struct {
	name, args string
}

var resourceTypes = []ResourceType{
	{
		"mysql",
		"--tier=db-g1-small,--activation-policy=ALWAYS,--db-version=MYSQL_5_7",
	},
	{
		"_postgres",
		"--cpu=1,--memory=3840MiB,--db-version=POSTGRES_9_6,--activation-policy=ALWAYS",
	},
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	stdcli.AddCommand(&cli.Command{
		Name:   "infra",
		Usage:  "Managed GCP stack resources and infrastructure",
		Action: cmdResourceList,
		Subcommands: []*cli.Command{
			{
				Name:            "create",
				Usage:           "create a new resource",
				Action:          cmdResourceCreate,
				Flags:           []cli.Flag{stackFlag},
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
	client := getClient(c)
	resp, err := client.Provider().ResourceList()
	if err != nil {
		return err
	}

	fmt.Printf("Name: %s\n", client.StackName)
	fmt.Printf("GCP Project: %s\n", client.ProjectId)
	fmt.Println("\nResource:")

	for _, r := range resp {
		kind := r.Kind.String()
		if _, err := checkResourceType(kind); err == nil {
			fmt.Printf("%s:%s\n", kind, r.Name)
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

	rs, err := getClient(c).GetResource(name)
	if err != nil {
		return err
	}

	fmt.Printf("%s ", rs.Name)
	for k, v := range rs.Exports {
		fmt.Printf("%s=%s ", k, v)
	}

	fmt.Printf("\n")
	return nil
}

func cmdResourceDelete(c *cli.Context) error {
	name := c.Args().Get(0)
	if len(name) == 0 {
		log.Errorf("no name given. Usage: \n")
		stdcli.Usage(c)
	}

	err := getClient(c).Provider().ResourceDelete(name)
	if err != nil {
		return err
	}

	fmt.Println("\nDELETED")
	return nil
}

func cmdResourceCreate(c *cli.Context) error {
	t, err := checkResourceType(c.Args().First())
	if err != nil {
		return err
	}

	args := append(strings.Split(t.args, ","), c.Args().Tail()...)
	stdcli.EnsureOnlyFlags(c, args)
	options := stdcli.FlagsToOptions(c, args)

	var optionsList []string
	for key, val := range options {
		optionsList = append(optionsList, fmt.Sprintf("%s=%q", key, val))
	}

	if options["name"] == "" {
		options["name"] = fmt.Sprintf("%s-%d", t.name, (rand.Intn(89999) + 1000))
	}

	fmt.Printf("Creating %s (%s", options["name"], t.name)
	if len(optionsList) > 0 {
		sort.Strings(optionsList)
		fmt.Printf(": %s", strings.Join(optionsList, " "))
	}
	fmt.Printf(")... ")
	fmt.Printf("\n")

	rs, err := getClient(c).CreateResource(t.name, options)
	if err != nil {
		return err
	}

	log.Debugf("Resource: %+v", rs)
	fmt.Println("\nCREATED")
	return nil
}

func cmdLinkCreate(c *cli.Context) error {
	_, app, err := getDirApp(".")
	if err != nil {
		return err
	}
	name := c.Args().First()

	err = getClient(c).CreateResourceLink(app, name)
	if err != nil {
		return err
	}

	fmt.Printf("Linked %s to %s\n", name, app)
	return nil
}

func cmdLinkDelete(c *cli.Context) error {
	_, app, err := getDirApp(".")
	if err != nil {
		return err
	}
	name := c.Args().First()

	err = getClient(c).DeleteResourceLink(app, name)
	if err != nil {
		return err
	}

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
