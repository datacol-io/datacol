package models

import (
	homedir "github.com/mitchellh/go-homedir"
	"log"
	"path/filepath"
	"time"
)

var (
	ConfigPath, ApprcPath string
	SvaFilename           = "sva.json"
	AwsCredentialFile     = "credentials.csv"
	AwsKeyPemPath         = "key.pem"
	SortableTime          = "20060102.150405.000000000"
)

const (
	StatusCreated = "created"
)

func init() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	ConfigPath = filepath.Join(home, ".datacol")
	ApprcPath = filepath.Join(ConfigPath, "config.json")
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

type AppCreateOptions struct {
	RepoUrl string
	Auth    map[string]string
}

type CreateBuildOptions struct {
	Version     string
	Auther      string
	Description string
}
