package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/common"
	docker "github.com/fsouza/go-dockerclient"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/storage/v1"
)

const (
	buildKind   = "Build"
	releaseKind = "Release"
)

func (g *GCPCloud) BuildGet(app, id string) (*pb.Build, error) {
	b, err := g.store.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	// Sometime GCP don't assign Status for a newly trigged build. We should also check for empty build status.
	if b.Status == "WORKING" || b.Status == "CREATED" || b.Status == "" {
		cb := g.cloudbuilder()
		rb, err := cb.Projects.Builds.Get(g.Project, b.RemoteId).Do()
		if err != nil {
			return nil, err
		}

		if b.Status != rb.Status {
			b.Status = rb.Status
			return b, g.store.BuildSave(b)
		}
	}

	return b, nil
}

func (g *GCPCloud) BuildDelete(app, id string) error {
	ctx, key := g.nestedKey(buildKind, id)
	return g.datastore().Delete(ctx, key)
}

func (g *GCPCloud) BuildList(app string, limit int64) (pb.Builds, error) {
	return g.store.BuildList(app, limit)
}

func (g *GCPCloud) BuildCreate(app string, req *pb.CreateBuildOptions) (*pb.Build, error) {
	id := req.DockerTag
	if id == "" {
		id = generateId("B", 5)
	}

	build := &pb.Build{
		App:       app,
		Id:        id,
		Status:    "CREATED",
		Version:   req.Version,
		Procfile:  req.Procfile,
		CreatedAt: timestampNow(),
	}

	return build, g.store.BuildSave(build)
}

func (a *GCPCloud) BuildImport(id string, tr io.Reader, w io.WriteCloser) error {
	build, err := a.BuildGet("", id)
	if err != nil {
		return err
	}
	dkr, err := docker.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("docker client: %v", err)
	}

	app, id := build.App, build.Id
	target := fmt.Sprintf("gcr.io/%s/%v", a.Project, app)

	err = common.BuildDockerLoad(target, id, dkr, tr, w, nil)
	if err == nil {
		build.Status = "SUCCESS"
	} else {
		build.Status = "FAILED"
	}

	// Ignore the error by build save
	if berr := a.store.BuildSave(build); berr != nil {
		log.Errorf("Failed to save build: %v", berr)
	}

	return err
}

func (g *GCPCloud) BuildUpload(id, filename string) error {
	build, err := g.BuildGet("", id)
	if err != nil {
		return err
	}

	service, bucket := g.storage(), g.BucketName
	gskey, app := fmt.Sprintf("%s.tar.gz", id), build.App

	log.Infof("Pushing code to gs://%s/%s", bucket, gskey)

	object := &storage.Object{
		Bucket:      bucket,
		Name:        gskey,
		ContentType: "application/gzip",
	}

	reader, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("reading tempfile err: %v", err)
	}
	defer reader.Close()

	if _, err := service.Objects.Insert(bucket, object).Media(reader).Do(); err != nil {
		return fmt.Errorf("Uploading to gs://%s/%s err: %s", bucket, gskey, err)
	}

	cb := g.cloudbuilder()

	log.Infof("Building from gs://%s/%s", bucket, gskey)
	tag := fmt.Sprintf("gcr.io/$PROJECT_ID/%v:%v", app, id)
	latestTag := fmt.Sprintf("gcr.io/$PROJECT_ID/%v:latest", app)

	stepBuildArgs := []string{"build", "-t", tag, "-t", latestTag}
	envVars, _ := g.EnvironmentGet(app)
	for key, val := range envVars {
		if val != "" {
			stepBuildArgs = append(stepBuildArgs, fmt.Sprintf("--build-arg=%s=%s", key, val))
		}
	}

	op, err := cb.Projects.Builds.Create(g.Project, &cloudbuild.Build{
		LogsBucket: bucket,
		Source: &cloudbuild.Source{
			StorageSource: &cloudbuild.StorageSource{
				Bucket: bucket,
				Object: gskey,
			},
		},
		Steps: []*cloudbuild.BuildStep{
			{
				Name: "gcr.io/cloud-builders/docker:17.12.0",
				Args: append(stepBuildArgs, "."),
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

	build.RemoteId = remoteId
	return g.store.BuildSave(build)
}

func (g *GCPCloud) BuildLogs(app, id string, index int) (int, []string, error) {
	storageService := g.storage()
	i, logs, err := buildLogs(storageService, g.BucketName, id, index)
	if err != nil {
		return i, logs, err
	}

	return i, logs, err
}

func (g *GCPCloud) BuildLogsStream(id string) (io.Reader, error) {
	return nil, fmt.Errorf("Not supported on GCP.")
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

func buildLogs(service *storage.Service, bucket, bid string, index int) (int, []string, error) {
	if bid == "" {
		return 0, []string{}, errors.New("GCR build Id (build.RemoteID) is not provided")
	}

	key := fmt.Sprintf("log-%s.txt", bid)
	log.Debugf("fetching logs from gs://%s/%s", bucket, bid)
	lines := []string{}

	// container builder might take little time to download source from storage bucket.
	// We will loop for 5 minutes and check the status after 1 second.

	timeout := time.After(5 * time.Minute)
	tick := time.Tick(1 * time.Second)
	var resp *http.Response
	var err error

Loop:
	for {
		select {
		case <-tick:
			resp, err = service.Objects.Get(bucket, key).Download()
			if err != nil {
				if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
					fmt.Print(".")
				} else {
					return index, lines, fmt.Errorf("fetching body err: %v", err)
				}
			} else {
				break Loop
			}
		case <-timeout:
			return index, lines, fmt.Errorf("Timeout for fetching logs. err: %s", err)
		}
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return index, lines, err
	}

	parts := strings.Split(string(body), "\n")

	for _, line := range parts[index:] {
		if len(line) > 0 && line != "\n" {
			lines = append(lines, line)
		}
	}

	return len(parts) - 1, lines, nil
}

func waitForOp(svc *cloudbuild.Service, stsvc *storage.Service, projectId, bucket, id string) (string, error) {
	log.Infof("Waiting on build %s", id)
	status := "PENDING"
	index := 0

	for {
		time.Sleep(2 * time.Second)
		b, err := svc.Projects.Builds.Get(projectId, id).Do()
		if err != nil {
			log.Fatal(err)
		}
		status = b.Status

		logKey := fmt.Sprintf("log-%s.txt", id)
		i, logs, err := buildLogs(stsvc, bucket, logKey, index)
		index = i
		if err != nil {
			log.Fatal(err)
		}

		for _, line := range logs {
			fmt.Println(line)
		}

		if b.Status != "WORKING" && b.Status != "QUEUED" {
			fmt.Printf("\n")
			log.Infof("Build status: %s", b.Status)
			break
		}
	}

	return status, nil
}
