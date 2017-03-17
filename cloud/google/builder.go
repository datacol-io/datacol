package google

import (
  "fmt"
  "bytes"
  "log"
  "encoding/json"
  "time"

  "google.golang.org/api/storage/v1"
  "google.golang.org/api/cloudbuild/v1"
  "google.golang.org/api/googleapi"
  "github.com/dinesh/rz/client/models"
)

func (g *GCPCloud) BuildImport(gskey string, tarf []byte) error {
  service := g.storage()
  bucket := g.BucketName

  fmt.Printf("Pushing code to gs://%s/%s\n", g.BucketName, gskey)

  object := &storage.Object{
    Bucket:       bucket,
    Name:         gskey,
    ContentType: "application/gzip",
  }

  if _, err := service.Objects.Insert(bucket, object).Media(bytes.NewBuffer(tarf)).Do(); err != nil { 
    return fmt.Errorf("Uploading to gs://%s/%s err: %s", bucket, gskey, err)
  }

  return nil
}

func (g *GCPCloud) BuildCreate(app string, gskey string, opts *models.BuildOptions) error {
  service := g.cloudbuilder()
  bucket  := g.BucketName

  fmt.Printf("Building from gs://%s/%s\n", bucket, gskey)
  tag := fmt.Sprintf("gcr.io/$PROJECT_ID/%v:%v", app, opts.Id)
  latestTag := fmt.Sprintf("gcr.io/$PROJECT_ID/%v:latest", app)

  op, err := service.Projects.Builds.Create(g.Project, &cloudbuild.Build{
    LogsBucket: bucket,
    Source: &cloudbuild.Source{
      StorageSource: &cloudbuild.StorageSource{
        Bucket: bucket,
        Object: gskey,
      },
    },
    Steps: []*cloudbuild.BuildStep{
      {
        Name: "gcr.io/cloud-builders/docker",
        Args: []string{"build", "-t", tag, "-t", latestTag, "."},
      },
    },
    Images: []string{tag},
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

  fmt.Printf("Logs at https://console.cloud.google.com/m/cloudstorage/b/%s/o/log-%s.txt\n", bucket, remoteId)

  return waitForOp(service, g.Project, remoteId)
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
      fmt.Printf("\nBuild status: %v\n", b.Status)
      break
    }
  }

  return nil
}
