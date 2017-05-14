package models

import (
	homedir "github.com/mitchellh/go-homedir"
	"log"
	"path/filepath"
	"time"
)

var (
	ConfigPath  string
	SvaFilename = "sva.json"
)

func init() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	ConfigPath = filepath.Join(home, ".datacol")
}

type Environment map[string]string

type Apps []*App
type Builds []*Build
type Releases []*Release
type Resources []*Resource

type ReleaseOptions struct {
	Port int
	Env  string
	Wait bool
}

type LogStreamOptions struct {
	Follow bool
	Since  time.Duration
}
