package modelss

import (
	homedir "github.com/mitchellh/go-homedir"
	"log"
	"path/filepath"
	"time"
)

var (
	ConfigPath  string
	SvaFilename string
)

func init() {
	SvaFilename = "sva.json"

	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err)
	}
	ConfigPath = filepath.Join(home, ".datacol")
}

type Stack struct {
	Name       string `datastore:"name"`
	ProjectId  string `datastore:"project_id"`
}

type App struct {
	Name     		string `datastore:"name"`
	Status   		string `datastore:"status"`
	ReleaseId  	string `datastore:"release_id"`
	HostPort 		string `datastore:"host_port"`
}

type Apps []*App

type Build struct {
	Id        string    `datastore:"id"`
	App       string    `datastore:"app"`
	RemoteId  string    `datastore:"remote_id,noindex"`
	Status    string    `datastore:"status,noindex"`
	CreatedAt time.Time `datastore:"created_at,noindex"`
}

type Builds []*Build

type Release struct {
	Id        string    `datastore:"id"`
	App       string    `datastore:"app"`
	BuildId   string    `datastore:"buildId"`
	Status    string    `datastore:"status"`
	CreatedAt time.Time `datastore:"created_at"`
}

type Releases []*Release

type ReleaseOptions struct {
	Port int
	Env string
	Wait bool
}

type BuildOptions struct {
	Id  string
	Key string
}

type Environment map[string]string

type LogStreamOptions struct {
	Follow bool
	Since  time.Duration
}

type ResourceVar struct{
	Name  string  `datastore:"name,noindex"`
	Value string 	`datastore:"value,noindex"`
}

type Resource struct {
	Name   string `datastore:"name"`
	Status string `datastore:"status,noindex"`
	Kind   string `datastore:"kind,noindex"`
	URL    string `datastore:"url,noindex"`

	Apps    []string          `datastore:"apps,noindex"`
	Exports []ResourceVar 		`datastore:"exports,noindex"`

	Parameters map[string]string `datastore:"-"`
	Outputs    map[string]string `datastore:"-"`
}

type Resources []Resource
