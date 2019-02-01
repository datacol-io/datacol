package models

import (
	"log"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
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
type ProcfileMap map[string]string

type Apps []*App
type Builds []*Build
type Releases []*Release
type Resources []*Resource
type AppDomains []string

type ReleaseOptions struct {
	Port   int
	Env    string
	Wait   bool
	Domain string
	Expose bool // should be expose the service to the public
}

type LogStreamOptions struct {
	Follow    bool
	Proctype  string
	Since     time.Duration
	TailLines string // number of recent lines to show for streaming (string format)
}

type AppCreateOptions struct {
	RepoUrl string
	Auth    map[string]string
}

type CreateBuildOptions struct {
	Version     string
	Author      string
	Description string
	Procfile    []byte
	DockerTag   string
	Trigger     bool // trigger the build in the cloud (in gcp-cloudbuild or aws-codebuild)
}

type ProcessRunOptions struct {
	Entrypoint    []string
	Tty           bool
	Detach        bool
	Width, Height int
}

type ProcessList []*Process
