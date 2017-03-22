package main

import (
  "bytes"
  "os"
  "os/exec"
  "syscall"
  "fmt"
  "path/filepath"
  log "github.com/Sirupsen/logrus"

  "github.com/dinesh/datacol/client"
  "github.com/dinesh/datacol/client/models"
  "github.com/dinesh/datacol/cmd/stdcli"
)

func cmdKubectl(args []string) {
  ct := client.Client{}
  if err := ct.SetStack(stdcli.GetStack()); err != nil {
    log.Fatal(err)
  }

  token, err := ct.Provider().CacheCredentials()
  if err != nil { log.Fatal(err) }

  excode := execute(ct.Stack.Name, token, args)
  os.Exit(excode)
}

func execute(env, token string, args []string) int {
  var (
    out, outErr bytes.Buffer
    exitcode int
  )

  cfgpath := filepath.Join(models.ConfigPath, env, "kubeconfig")
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