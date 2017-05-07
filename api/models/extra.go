package models

import (
  "time"
  "log"
  "path/filepath"
  homedir "github.com/mitchellh/go-homedir"
)

var (
  ConfigPath string
  SvaFilename = "sva.json"
)

func init(){
  home, err := homedir.Dir()
  if err != nil { log.Fatal(err) }
  ConfigPath = filepath.Join(home, ".datacol")

}

type Apps []*App
type Builds []*Build
type Releases []*Release
type Resources []*Resource

type BuildOptions struct {
  Id  string
  Key string
}

type Environment map[string]string

type ReleaseOptions struct {
  Port int
  Env string
  Wait bool
}

type LogStreamOptions struct {
  Follow bool
  Since  time.Duration
}
