package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"gopkg.in/urfave/cli.v2"

	"github.com/dinesh/datacol/client"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cmd/stdcli"
)

func init() {
	stdcli.AddCommand(&cli.Command{
		Name:            "kubectl",
		Usage:           "kubectl wrapper for datacol",
		Action:          cmdKubectl,
		SkipFlagParsing: true,
	})
}

func cmdKubectl(c *cli.Context) error {
	ct := client.Client{}
	ct.SetFromEnv()

	excode := execute(ct.StackName, c.Args().Slice())
	os.Exit(excode)
	return nil
}

func execute(env string, args []string) int {
	var (
		out, outErr bytes.Buffer
		exitcode    int
	)

	cfgpath := filepath.Join(pb.ConfigPath, env, "kubeconfig")
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
