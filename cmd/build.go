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

  "github.com/dinesh/rz/cmd/stdcli"
  "github.com/dinesh/rz/client"
  provider "github.com/dinesh/rz/cloud/google"
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
  dir := "."
  if c.NArg() > 0 {
    dir = c.Args().Get(0)
  }

  _, name, err := getDirApp(dir)
  if err != nil { return err }

  app, err := client.GetApp(name)
  if err != nil { return err }

  build := client.NewBuild(app)
  return executeBuildDir(c, build, dir)
}

func executeBuildDir(c *cli.Context, b *client.Build, dir string) error {
  fmt.Println("Creating tarball ...")

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

func uploadBuildSource(c *cli.Context, b *client.Build, tarf []byte) (string, error) {
  client := getClient(c)
  bucket := client.Stack.Bucket
  objectName := fmt.Sprintf("%s.tar.gz", b.Id)

  if err := provider.UploadSource(client.PrdClient(), bucket, objectName, tarf); err != nil {
    return "", nil
  }
  return objectName, nil
}

func finishBuild(c *cli.Context, b *client.Build, objectName string) error {
  client := getClient(c)
  bopts := &provider.BuildOpts{
    ProjectId:  client.Stack.ProjectId,
    Bucket:     client.Stack.Bucket,
    ObjectName: objectName,
  }

  return provider.BuildWithGCR(client.PrdClient(), b.AppName, bopts)
}


