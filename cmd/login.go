package main

import (
  "bufio"
  "os"
  log "github.com/Sirupsen/logrus"
  "golang.org/x/net/context"
  "fmt"
  pbs "github.com/dinesh/datacol/api/controller"
  "github.com/dinesh/datacol/cmd/stdcli"
  "gopkg.in/urfave/cli.v2"
)

func init() {
  stdcli.AddCommand(&cli.Command{
    Name:   "login",
    Usage:  "login",
    Action: cmdLogin,
  })
}

func cmdLogin(c *cli.Context) error {
  r := bufio.NewReader(os.Stdin)
  fmt.Print("Enter your password: ")

  passwd, err := r.ReadString('\n')
  if err != nil {
    log.Fatal(err)
  }

  api, close := getApiClient(c)
  defer close()

  ret, err := api.Auth(context.TODO(), &pbs.AuthRequest{Password: passwd})
  if err != nil {
    return err
  }

  log.Debugf("response: %+v", ret)
  fmt.Printf("Successfully Logged in.")
  return nil
}
