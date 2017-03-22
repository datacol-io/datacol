package main

import (
  "fmt"
  "os"
  "path"
  "path/filepath"
  "io/ioutil"

  "github.com/docker/docker/builder/dockerignore"
  "github.com/docker/docker/pkg/fileutils"
  "github.com/docker/docker/pkg/archive"
  "gopkg.in/urfave/cli.v2"

  "github.com/dinesh/datacol/cmd/stdcli"
  "github.com/dinesh/datacol/client/models"
  "github.com/dinesh/datacol/client"
)

func init(){
  stdcli.AddCommand(cli.Command{
    Name: "build",
    UsageText: "build an app",
    Action: cmdBuild,
  })
}

func cmdBuild(c *cli.Context) error {
  client := getClient(c)

  dir, name, err := getDirApp(".")
  if err != nil { return err }

  app, err := client.GetApp(name)
  if err != nil { 
    return app404Err(name)
  }

  build := client.NewBuild(app)
  return executeBuildDir(c, build, dir)
}

func executeBuildDir(c *cli.Context, b *models.Build, dir string) error {
  fmt.Print("Creating tarball ...")

  tar, err := createTarball(dir)
  if err != nil { return err }

  fmt.Println("OK")

  objectName, err := uploadBuildSource(c, b, tar)
  if err != nil { 
    return err
  }

  return finishBuild(c, b, objectName)
}

func createTarball(base string) ([]byte, error) {
  cwd, err := os.Getwd()
  if err != nil {
    return nil, err
  }

  sym, err := filepath.EvalSymlinks(base)
  if err != nil {
    return nil, err
  }

  err = os.Chdir(sym)
  if err != nil {
    return nil, err
  }

  var includes = []string{"."}
  var excludes []string

  dockerIgnorePath := path.Join(sym, ".dockerignore")
  dockerIgnore, err := os.Open(dockerIgnorePath)
  if err != nil {
    if !os.IsNotExist(err) {
      return nil, err
    }
    excludes = make([]string, 0)
  } else {
    excludes, err = dockerignore.ReadAll(dockerIgnore)
    if err != nil {
      return nil, err
    }
  }

  keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
  keepThem2, _ := fileutils.Matches("Dockerfile", excludes)
  if keepThem1 || keepThem2 {
    includes = append(includes, ".dockerignore", "Dockerfile")
  }

  options := &archive.TarOptions{
    Compression:     archive.Gzip,
    ExcludePatterns: excludes,
    IncludeFiles:    includes,
  }

  out, err := archive.TarWithOptions(sym, options)
  if err != nil {
    return nil, err
  }

  bytes, err := ioutil.ReadAll(out)
  if err != nil {
    return nil, err
  }

  err = os.Chdir(cwd)
  if err != nil {
    return nil, err
  }

  return bytes, nil
}

func uploadBuildSource(c *cli.Context, b *models.Build, tarf []byte) (string, error) {
  client := getClient(c)
  source := fmt.Sprintf("%s.tar.gz", b.Id)

  if err := client.Provider().BuildImport(source, tarf); err != nil {
    return "", nil
  }
  return source, nil
}

func finishBuild(c *cli.Context, b *models.Build, objectName string) error {
  bopts := &models.BuildOptions{Key: objectName, Id: b.Id}
  
  err := getClient(c).Provider().BuildCreate(b.App, objectName, bopts)
  if err != nil {
    b.Status = "failed"
  } else {
    b.Status = "success"
    if err := client.Persist([]byte("builds"), b.Id, b); err != nil {
      return err
    }
  }
  
  return err
}


