package client

import (
  "os"
  "fmt"
  "context"
  "log"
  "path"
  "bytes"
  "io/ioutil"
  "time"
  "path/filepath"
  "encoding/json"

  "google.golang.org/api/storage/v1"
  "google.golang.org/api/cloudbuild/v1"
  "golang.org/x/oauth2/google"
  "github.com/docker/docker/builder/dockerignore"
  "github.com/docker/docker/pkg/archive"
  "github.com/docker/docker/pkg/fileutils"
)

var (
  projectId string
  bucket string
)

func ExecuteBuildDir(projectId, bucket, dir string) error {
  dir, err := filepath.Abs(dir)
  if err != nil { return err }

  log.Printf("Creating tarball... ")

  tar, err := createTarball(dir)
  if err != nil { return err }

  log.Printf("OK")
  name, err := uploadBuildSource(bucket, projectId, tar)
  if err != nil { return err }

  return finishBuild(projectId, bucket, name)  
}

func uploadBuildSource(bucket string, projectId string, tarf []byte) (string, error) {
  hc, err := google.DefaultClient(context.TODO())
  if err != nil { return "", err }

  buildId := generateId("B", 10)
  objectName  := fmt.Sprintf("%s.tar.gz", buildId)
  log.Printf("Pushing code to gs://%s/%s", bucket, objectName)
  
  service, err := storage.New(hc)
  if err != nil { return "", err }
  
  object := &storage.Object{
    Bucket: bucket,
    Name: objectName,
    ContentType: "application/gzip",
  }

  if _, err = service.Objects.Insert(bucket, object).Media(bytes.NewBuffer(tarf)).Do(); err != nil { 
    return "", err 
  }

  return objectName, nil
}

func finishBuild(projectId, bucket, objectName string) error {
  hc, err := google.DefaultClient(context.TODO())
  if err != nil { return err }

  service, err := cloudbuild.New(hc)
  if err != nil { return err }

  appName := generateId("b-", 10)
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

  if err != nil { return err }

  remoteId, err := getBuildID(op)
  if err != nil { return err }

  awaitCSOp(service, projectId, remoteId)
  log.Printf("Logs at https://console.cloud.google.com/m/cloudstorage/b/%s/o/log-%s.txt", bucket, remoteId)

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