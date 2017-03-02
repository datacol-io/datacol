package build

import (
  "os"
  "fmt"
  "log"
  "path"
  "bytes"
  "io/ioutil"
  "time"
  "path/filepath"
  "encoding/json"

  "github.com/pkg/errors"
  "google.golang.org/api/storage/v1"
  "google.golang.org/api/cloudbuild/v1"
  "github.com/docker/docker/builder/dockerignore"
  "github.com/docker/docker/pkg/archive"
  "github.com/docker/docker/pkg/fileutils"
  "google.golang.org/api/googleapi"

  "github.com/dinesh/rz/client/helper"
  "github.com/dinesh/rz/client"
)

func ExecuteBuildDir(app *client.App, auth *client.Auth) error {
  dir := app.Dir

  dir, err := filepath.Abs(dir)
  if err != nil { return err }

  log.Printf("Creating tarball...")

  tar, err := createTarball(dir)
  if err != nil { return err }

  log.Printf("OK")
  buildId, err := uploadBuildSource(tar, app, auth)
  if err != nil { return err }

  return finishBuild(buildId, app, auth)
}

func uploadBuildSource(tarf []byte, app *client.App, auth *client.Auth) (string, error) {
  bucket  := auth.BucketName
  hc := client.GetClientOrDie(auth)

  buildId := helper.GenerateId("B", 10)
  objectName  := fmt.Sprintf("%s.tar.gz", buildId)
  log.Printf("Pushing code to gs://%s/%s", bucket, objectName)
  
  service, err := storage.New(hc)
  if err != nil { 
    return "", errors.Wrap(err, "failed to get storage client") 
  }

  object := &storage.Object{
    Bucket:       bucket,
    Name:         objectName,
    ContentType: "application/gzip",
  }

  if _, err = service.Objects.Insert(bucket, object).Media(bytes.NewBuffer(tarf)).Do(); err != nil { 
    return "", errors.Wrap(err, fmt.Sprintf("failed to upload gs://%s/%s", bucket, objectName))
  }

  return objectName, nil
}

func CancelBuild(auth *client.Auth, buildId string) error {
  hc := client.GetClientOrDie(auth)
  service, err := cloudbuild.New(hc)
  if err != nil { return errors.Wrap(err, "failed to get cloudbuild client") }
  
  if _, err = service.Projects.Builds.Cancel(auth.ProjectId, buildId, &cloudbuild.CancelBuildRequest{}).Do(); err != nil {
    return err
  }

  return nil
}

func finishBuild(objectName string, app *client.App, auth *client.Auth) error {
  projectId := auth.ProjectId
  bucket    := auth.BucketName
  appName   := app.Name
  hc        := client.GetClientOrDie(auth)

  service, err := cloudbuild.New(hc)
  if err != nil { return errors.Wrap(err, "failed to get cloudbuild client") }

  log.Printf("Building from gs://%s/%s", bucket, objectName)

  op, err := service.Projects.Builds.Create(projectId, &cloudbuild.Build{
    LogsBucket: bucket,
    Source: &cloudbuild.Source{
      StorageSource: &cloudbuild.StorageSource{
        Bucket: bucket,
        Object: objectName,
      },
    },
    Steps: []*cloudbuild.BuildStep{
      {
        Name: "gcr.io/cloud-builders/dockerizer",
        Args: []string{"gcr.io/" + projectId + "/" + appName },
      },
    },
    Images: []string{"gcr.io/" + projectId + "/" + appName },
  }).Do()

  if err != nil {
    if ae, ok := err.(*googleapi.Error); ok && ae.Code == 403 {
      log.Fatal(ae)
    }
    return errors.Wrap(err, "failed to initiate build")
  }

  remoteId, err := getBuildID(op)
  if err != nil { return errors.Wrap(err, "failed to get Id for build") }
  log.Printf("Logs at https://console.cloud.google.com/m/cloudstorage/b/%s/o/log-%s.txt", bucket, remoteId)

  awaitCSOp(service, projectId, remoteId)

  return nil
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

func awaitCSOp(svc *cloudbuild.Service, projectId string, id string) error {
  log.Printf("Waiting on build %s\n", id)

  for {
    time.Sleep(2 * time.Second)
    b, err := svc.Projects.Builds.Get(projectId, id).Do()
    if err != nil { return err }
    
    if b.Status != "WORKING" && b.Status != "QUEUED" {
      log.Printf("Build status: %v\n", b.Status)
      break
    }
  }

  return nil
}

func getBuildID(op *cloudbuild.Operation) (string, error) {
  if len(op.Metadata) == 0 {
    return "", fmt.Errorf("missing Metadata in operation")
  }
  
  bm := &cloudbuild.BuildOperationMetadata{}
  if err := json.Unmarshal(op.Metadata, &bm); err != nil {
    return "", err
  }
  
  return bm.Build.Id, nil
}
