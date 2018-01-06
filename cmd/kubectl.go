package cmd

import (
	"fmt"
	"os"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	pbs "github.com/dinesh/datacol/api/controller"
	"github.com/dinesh/datacol/cmd/stdcli"
	"github.com/urfave/cli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:            "kubectl",
		Usage:           "kubectl wrapper for datacol",
		Action:          cmdKubectl,
		SkipFlagParsing: true,
	})
}

func cmdKubectl(c *cli.Context) error {
	client, close := getApiClient(c)
	defer close()

	args := c.Args()
	ret, err := client.ProviderServiceClient.Kubectl(
		context.TODO(), &pbs.KubectlReq{Args: args},
	)
	stdcli.ExitOnError(err)

	onApiExec(ret, args)
	return nil
}

func onApiExec(ret *pbs.CmdResponse, args []string) {
	exitcode := int(ret.ExitCode)
	if len(ret.Err) > 0 {
		log.Warn(ret.Err)
		log.Warn(ret.StdErr)
		fmt.Printf("failed to execute %v\n", args)
	} else {
		if exitcode == 0 {
			fmt.Printf(ret.StdOut)
		} else {
			fmt.Printf(ret.StdErr)
		}
	}

	os.Exit(exitcode)
}
