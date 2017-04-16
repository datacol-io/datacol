package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"gopkg.in/urfave/cli.v2"

	"github.com/dinesh/datacol/client"
	"github.com/dinesh/datacol/client/models"
	"github.com/dinesh/datacol/cmd/stdcli"
)

func init() {
	stdcli.AddCommand(cli.Command{
		Name:            "kubectl",
		Usage:           "kubectl wrapper for datacol",
		Action:          cmdKubectl,
		SkipFlagParsing: true,
	})
}

func cmdKubectl(c *cli.Context) {
	ct := client.Client{}
	if err := ct.SetStack(stdcli.GetStack()); err != nil {
		log.Fatal(err)
	}

	excode := execute(ct.Stack.Name, c.Args())
	os.Exit(excode)
}

func execute(env string, args []string) int {
	var (
		out, outErr bytes.Buffer
		exitcode    int
	)

	cfgpath := filepath.Join(models.ConfigPath, env, "kubeconfig")
	args = append([]string{"--kubeconfig", cfgpath, "-n", env}, args...)
	c := exec.Command("kubectl", args...)

	c.Stdout = &out
	c.Stderr = &outErr
	err := c.Run()

	if exitError, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
			exitcode = waitStatus.ExitStatus()
		}
	}
	if exitcode == 0 {
		fmt.Printf(out.String())
	} else {
		fmt.Printf(outErr.String())
	}

	return exitcode
}
