package client

import (
  "os"
  "fmt"
  "io/ioutil"
  "log"
  "encoding/json"
  "path/filepath"

   homeDir "github.com/mitchellh/go-homedir"
)

var (
  home, _   = homeDir.Dir()
  apprcPath = home + "/.razorbox/rzrc.json"
)

type App struct {
  Name string `json: "name"`
  Dir  string `json: "dir"`
}

type Auth struct {
  RackName   string  `json: "rack_name"`
  ServiceKey []byte  `json: "service_key"`
  ProjectId  string  `json: "project_id"`
  Zone       string  `json: "zone.omitempty"`
  BucketName string  `json: "bucket,omitempty"`
  Apps       []*App  `json: "apps"`
}

type Apprc struct {
  Context string  `json: "context"`
  Auths  []*Auth  `json: "auths"`
}

/* Exits if there is any error.*/
func (rc *Apprc) GetAuth() *Auth {
  if rc.Context != "" {
    for _, a := range rc.Auths {
      if a.RackName == rc.Context {
        return a
      }
    }
  }
  return nil
}

func (rc *Apprc) SetAuth(a *Auth) error {
  for i, b := range rc.Auths {
    if b.RackName == a.RackName {
      rc.Auths = append(rc.Auths[:i], rc.Auths[i+1:]...)
      break
    }
  }
  rc.Context = a.RackName
  rc.Auths = append(rc.Auths, a)
  return rc.Write()
}

func (rc *Apprc) DeleteAuth() error {
  if rc.Context != "" {
    for i, a := range rc.Auths {
      if a.RackName == rc.Context {
        rc.Auths = append(rc.Auths[:i], rc.Auths[i+1:]...)
        rc.Context = ""
        break
      }
    }
  }
  return rc.Write()
}

func(rc *Auth) GetApp(appName string) (*App, error) {
  if appName == "" {
    dirName, err := os.Getwd()
    if err != nil { return nil, err }

    appName = filepath.Base(dirName)
  }

  for _, a := range rc.Apps {
    if a.Name == appName {
      return a, nil
    }
  }

  return nil, fmt.Errorf("app not found: %s", appName)
}

func (auth *Auth) DeleteApp(appName string) error {
  for i, a := range auth.Apps {
    if a.Name == appName {
      auth.Apps = append(auth.Apps[:i], auth.Apps[i+1:]...)
    }
  }
  
  return nil
}

func (rc *Auth) SetApp(app *App) {
  for _, b := range rc.Apps {
    if b.Name == app.Name && app.Dir == b.Dir {
      return
    }
  }

  rc.Apps = append(rc.Apps, app)
}


func (rc *Apprc) Write() error {
  bytes, err := json.Marshal(rc)
  if err != nil { return err }

  return ioutil.WriteFile(apprcPath, bytes, 0600)
}

func LoadApprc() (*Apprc, error) {
  if _, err := os.Stat(apprcPath); err != nil {
    return nil, err
  }

  os.Chmod(apprcPath, 0600)
  rc := &Apprc{}

  bytes, err := ioutil.ReadFile(apprcPath)
  if err != nil { return nil, err }

  if err := json.Unmarshal(bytes, rc); err != nil { 
    return nil, err
  }

  return rc, nil
}

func EnsureApprc() error {
  if _, err := os.Stat(apprcPath); os.IsNotExist(err) {
    if err := os.MkdirAll(filepath.Dir(apprcPath), os.ModePerm); err != nil {
      return err
    }

    rc := &Apprc{}
    return rc.Write()
  }

  return nil
}

/* Exits if there is any error.*/
func GetAuthOrDie() (*Apprc, *Auth) {
  rc, err := LoadApprc()
  if err != nil {
    log.Fatal("Command requires razorbox initialization, please run `rz login`")
  }
  a := rc.GetAuth()
  if a == nil {
    log.Fatal("Not able to find valid rack name, please run `rz login`")
  }

  return rc, a
}

func SetAuth(a *Auth) error {
  rc, err := LoadApprc()
  if err != nil {
    rc = &Apprc{}
    rc.Auths = make([]*Auth, 0)
  }
  return rc.SetAuth(a)
}

func SetApp(appDir string) (*Auth, *App, error) {
  rc, auth := GetAuthOrDie()
  app  := &App{Name: filepath.Base(appDir), Dir: appDir }
  auth.SetApp(app)
  if err := rc.Write(); err != nil { return nil, nil, err }

  return auth, app, nil
}
