package client 

import (
  "path/filepath"
  "os"
  "time"
  "fmt"
  "net/http"

  "github.com/boltdb/bolt"
  homedir "github.com/mitchellh/go-homedir"
  "github.com/dinesh/rz/cmd/stdcli"
  provider "github.com/dinesh/rz/cloud/google"
)

var (
  rzDirName  = ".razorbox"
  dbFileName = "rz.db"
)

func init() {
  home, err := homedir.Dir()
  if err != nil {
    stdcli.Error(err)
    return
  }

  root := filepath.Join(home, rzDirName)
  
  if _, err = os.Stat(root); err != nil {
    if !os.IsNotExist(err) {
      stdcli.Error(err)
      return
    } else {
      if err := os.MkdirAll(root, 0700); err != nil {
        stdcli.Error(err)
        return
      }
    }
  }

  dbpath := filepath.Join(home, rzDirName, dbFileName)
  db, err := bolt.Open(dbpath, 0700, &bolt.Options{Timeout: 1 * time.Second})
  if err != nil {
    stdcli.Error(fmt.Errorf("creating database file: %v", err))
    return
  }
  
  DB = &Database{store: db}
}

type Client struct {
  Version     string
  StackName   string
  *Stack
}

func (c *Client) configRoot() string {
  home, _ := homedir.Dir()
  return filepath.Join(home, rzDirName, c.StackName)
}

func (c *Client) SetStack(name string) error {
  c.StackName = name
  st, err := findStack(name)
  if err != nil {
    return fmt.Errorf("Please create a stack with: $ rz init")
  }
  c.Stack = st

  return nil
}

func (c *Client) PrdClient() *http.Client {
  return provider.JwtClient(c.Stack.ServiceKey)
}

func (c *Client) PrdToken() (string, error) {
  return provider.BearerToken(c.Stack.ServiceKey)
}
