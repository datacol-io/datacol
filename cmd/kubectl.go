package main

import (
  "bytes"
  "os"
  "os/exec"
  "syscall"
  "fmt"
  "log"
  "path/filepath"

  "github.com/dinesh/rz/client"
  "github.com/dinesh/rz/cmd/stdcli"
  homeDir "github.com/mitchellh/go-homedir"
)

func cmdKubectl(args []string) {
  ct := client.Client{}
  if err := ct.SetStack(stdcli.GetStack()); err != nil {
    log.Fatal(err)
  }

  token, err := ct.Provider().BearerToken()
  if err != nil {
    log.Fatal(err)
  }

  excode := execute(ct.Stack.Name, token, args)
  os.Exit(excode)
}

func execute(env, token string, args []string) int {
  var (
    out, outErr bytes.Buffer
    exitcode int
  )

  home, _ := homeDir.Dir()
  cfgpath := filepath.Join(home, ".razorbox", env, "kubeconfig")
  
  args = append([]string{"--kubeconfig", cfgpath, "-n", env, "--token", token}, args...)
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