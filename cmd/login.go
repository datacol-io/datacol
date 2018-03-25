package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go/term"
	pbs "github.com/datacol-io/datacol/api/controller"
	"github.com/datacol-io/datacol/client"
	"github.com/datacol-io/datacol/cmd/stdcli"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:      "login",
		ArgsUsage: "[host]",
		Usage:     "login in to datacol",
		Action:    cmdLogin,
	})
}

func cmdLogin(c *cli.Context) error {
	host := c.Args().First()
	if host == "" {
		term.Warningln("Missing required argument: host")
		stdcli.Usage(c)
	}

	r := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your password: ")

	p, err := r.ReadString('\n')
	stdcli.ExitOnError(err)

	passwd := strings.TrimSpace(p)
	api, close := client.GrpcClient(host, passwd)
	defer close()

	log.Debugf("Trying to login with [%s]", passwd)
	ret, err := api.Auth(context.TODO(), &pbs.AuthRequest{Password: passwd})
	stdcli.ExitOnError(err)

	log.Debugf("response: %+v", toJson(ret))

	if err = updateSetting(ret, host, passwd); err != nil {
		stdcli.ExitOnError(err)
		return err
	}

	fmt.Printf("Successfully Logged in.")
	return nil
}

func updateSetting(ret *pbs.AuthResponse, host, passwd string) error {
	if err := createStackDir(ret.Name); err != nil {
		stdcli.ExitOnError(err)
		return err
	}

	if ret.Project == "" {
		return dumpAwsAuthParams(ret.Name, ret.Region, "", host, passwd)
	} else {
		return dumpGcpAuthParams(ret.Name, ret.Project, "", host, passwd)
	}
}
