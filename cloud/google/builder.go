package google

import (
  "fmt"
  "bytes"
  "log"
  "encoding/json"
  "time"
  "net/http"

  "google.golang.org/api/storage/v1"
  "google.golang.org/api/cloudbuild/v1"
  "google.golang.org/api/googleapi"
)

func UploadSource(htc *http.Client, bucket string, objectName string, tarf []byte) error {
  fmt.Printf("Pushing code to gs://%s/%s\n", bucket, objectName)

  service, err := storage.New(htc)
  
  if err != nil { 
    return fmt.Errorf("storage client %s", err) 
  }

  object := &storage.Object{
    Bucket:       bucket,
    Name:         objectName,
    ContentType: "application/gzip",
  }

  if _, err = service.Objects.Insert(bucket, object).Media(bytes.NewBuffer(tarf)).Do(); err != nil { 
    return fmt.Errorf("Uploading to gs://%s/%s err: %s", bucket, objectName, err)
  }

  return nil
}

type BuildOpts struct {
  ProjectId, Bucket, ObjectName string
}

func BuildWithGCR(htc *http.Client, appName string, opt *BuildOpts) error {
  service, err := cloudbuild.New(htc)
  if err != nil { 
    return fmt.Errorf("cloudbuilder client %v", err)
  }

  fmt.Printf("Building from gs://%s/%s\n", opt.Bucket, opt.ObjectName)

  op, err := service.Projects.Builds.Create(opt.ProjectId, &cloudbuild.Build{
    LogsBucket: opt.Bucket,
    Source: &cloudbuild.Source{
      StorageSource: &cloudbuild.StorageSource{
        Bucket: opt.Bucket,
        Object: opt.ObjectName,
      },
    },
    Steps: []*cloudbuild.BuildStep{
      {
        Name: "gcr.io/cloud-builders/dockerizer",
        Args: []string{"gcr.io/" + opt.ProjectId + "/" + appName },
      },
    },
    Images: []string{"gcr.io/" + opt.ProjectId + "/" + appName },
  }).Do()

  if err != nil {
    if ae, ok := err.(*googleapi.Error); ok && ae.Code == 403 {
      log.Fatal(ae)
    }

    return fmt.Errorf("failed to initiate build %v", err)
  }

  remoteId, err := getBuildID(op)
  if err != nil {
    return fmt.Errorf("failed to get Id for build %v", err) 
  }

  log.Printf("Logs at https://console.cloud.google.com/m/cloudstorage/b/%s/o/log-%s.txt\n", opt.Bucket, remoteId)

  return waitForOp(service, opt.ProjectId, remoteId)
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

func waitForOp(svc *cloudbuild.Service, projectId string, id string) error {
  fmt.Printf("Waiting on build %s\n", id)

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
