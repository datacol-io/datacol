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
	dkrconfig "github.com/docker/docker/cliconfig"
	dkrtypes "github.com/docker/engine-api/types"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:      "login",
		ArgsUsage: "[host]",
		Usage:     "login in to datacol",
		Action:    cmdLogin,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "host",
				Usage:  "datacol API host",
				EnvVar: "DATACOL_API_HOST",
			},
			cli.StringFlag{
				Name:   "api-key",
				Usage:  "datacol API Key",
				EnvVar: "DATACOL_API_KEY",
			},
		},
	})

	stdcli.AddCommand(cli.Command{
		Name:   "docker-login",
		Action: cmdDockerLogin,
	})
}

func cmdDockerLogin(c *cli.Context) error {
	client, close := getApiClient(c)
	defer close()

	result, err := client.GetDockerCreds()
	stdcli.ExitOnError(err)

	configfile, err := dkrconfig.Load("")
	stdcli.ExitOnError(err)
	configfile.AuthConfigs[result.Host] = dkrtypes.AuthConfig{
		Username:      result.Username,
		Password:      result.Password,
		ServerAddress: result.Host,
	}

	stdcli.ExitOnError(configfile.Save())
	return nil
}

func cmdLogin(c *cli.Context) error {
	host := c.Args().First()
	if host == "" {
		host = c.String("host")
		if host == "" {
			term.Warningln("Missing required argument: host or set DATACOL_API_HOST in env")
			stdcli.Usage(c)
		}
	}

	passwd := c.String("api-key")
	if passwd == "" {
		r := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your password: ")

		p, err := r.ReadString('\n')
		stdcli.ExitOnError(err)

		passwd = strings.TrimSpace(p)
	}

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
