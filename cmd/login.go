package main

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	pbs "github.com/dinesh/datacol/api/controller"
	"github.com/dinesh/datacol/client"
	"github.com/dinesh/datacol/cmd/stdcli"
	"golang.org/x/net/context"
	"gopkg.in/urfave/cli.v2"
	"os"
	"strings"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:   "login",
		Usage:  "login in to datacol",
		Action: cmdLogin,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "ip",
				Usage: "datacol stack hostname or IP address",
			},
		},
	})
}

func cmdLogin(c *cli.Context) error {
	stdcli.CheckFlagsPresence(c, "ip")

	r := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your password: ")

	p, err := r.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	passwd := strings.TrimSpace(p)
	host := c.String("ip")

	api, close := client.GrpcClient(host, passwd)
	defer close()

	log.Debugf("Trying to login with [%s]", passwd)
	ret, err := api.Auth(context.TODO(), &pbs.AuthRequest{Password: passwd})
	if err != nil {
		return err
	}

	log.Debugf("response: %+v", toJson(ret))
	if err = updateSetting(ret, host, passwd); err != nil {
		return err
	}

	fmt.Printf("Successfully Logged in.")
	return nil
}

func updateSetting(ret *pbs.AuthResponse, host, passwd string) error {
	if err := createStackDir(ret.Name); err != nil {
		return err
	}

	if ret.Project == "" {
		return dumpAwsAuthParams(ret.Name, ret.Region, host, passwd)
	} else {
		return dumpGcpAuthParams(ret.Name, ret.Project, "", host, passwd)
	}
}
