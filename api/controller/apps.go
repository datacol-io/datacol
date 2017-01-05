package controller

import (
  "fmt"
  "log"
  "net/http"
  "mime/multipart"
  "context"
  "io"

  "github.com/satori/go.uuid"
  "golang.org/x/oauth2/google"
  cstorage "cloud.google.com/go/storage"
  "google.golang.org/api/cloudbuild/v1"
  "google.golang.org/api/storage/v1"

  "github.com/dinesh/rz/api/domain"
)

func AppIndex(w http.ResponseWriter, r *http.Request){
  w.Write([]byte("AppIndex"))
}

func AppBuildCreate(w http.ResponseWriter, r *http.Request){
  ctx := context.Background()
  hc, err := google.DefaultClient(ctx, storage.CloudPlatformScope)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  tarf, _, err := r.FormFile("source")
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  prid := r.FormValue("project_id")
  if prid == "" {
    http.Error(w, "Required param missing: project_id", http.StatusBadRequest)
    return
  }

  appName := r.FormValue("app")
  if appName == "" {
    http.Error(w, "Required param missing: app", http.StatusBadRequest)
    return
  }

  bucketName  := "build"
  build       := domain.Build{ProjectId: prid, Id: fmt.Sprintf("%s",uuid.NewV4()), AppName: appName }
  buildObject := fmt.Sprintf("%s.tar.gz", build.Id)

  log.Printf("Pushing code to gs://%s/%s", bucketName, buildObject)

  if err := uploadTar(ctx, hc, bucketName, buildObject, build.ProjectId, tarf); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  service, err := cloudbuild.New(hc)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  call := service.Projects.Builds.Create(build.ProjectId, &cloudbuild.Build{
    LogsBucket: bucketName,
    Source: &cloudbuild.Source{
      StorageSource: &cloudbuild.StorageSource{
        Bucket: bucketName,
        Object: buildObject,
      },
    },
    Steps: []*cloudbuild.BuildStep{
      {
        Name: "gcr.io/cloud-builders/dockerizer",
        Args: []string{"gcr.io/" + build.ProjectId + "/" + build.AppName},
      },
    },
    Images: []string{"gcr.io/" + build.ProjectId + "/" + build.AppName},
  })

  _, err = call.Context(ctx).Do()
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  fmt.Fprintf(w, "Done")
}

func setupBucket(ctx context.Context, service *cstorage.Client, bucket string, projectID string) error {
  bkt := service.Bucket(bucket)

  if _, err := bkt.Attrs(ctx); err != nil {
    if err := bkt.Create(ctx, projectID, nil); err != nil {
      return err
    }
  }
  return nil
}

func uploadTar(ctx context.Context, client *http.Client, bucket string, objectName string, projectID string, tarf multipart.File) error {
  service, err := cstorage.NewClient(ctx)
  if err != nil {
    return err
  }

  err = setupBucket(ctx, service, bucket, projectID)
  if err != nil {
    return err
  }

  object := service.Bucket(bucket).Object(objectName)
  w := object.NewWriter(ctx)

  if _, err = io.Copy(w, tarf); err != nil {
    w.Close()
    return err
  }

  err = w.Close()
  return err
}
