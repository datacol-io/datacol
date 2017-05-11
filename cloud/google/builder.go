package google

import (
	"bytes"
	"cloud.google.com/go/datastore"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/storage/v1"
	"k8s.io/client-go/pkg/util/intstr"

	pb "github.com/dinesh/datacol/api/models"
)

const (
	buildKind   = "Build"
	releaseKind = "Release"
)

func (g *GCPCloud) BuildGet(app, id string) (*pb.Build, error) {
	var b pb.Build
	if err := g.datastore().Get(context.TODO(), g.nestedKey(buildKind, id), &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (g *GCPCloud) BuildDelete(app, id string) error {
	return g.datastore().Delete(context.TODO(), g.nestedKey(buildKind, id))
}

func (g *GCPCloud) BuildList(app string, limit int) (pb.Builds, error) {
	q := datastore.NewQuery(buildKind).
		Ancestor(g.stackKey()).
		Filter("App = ", app).
		Limit(limit)

	var builds pb.Builds
	_, err := g.datastore().GetAll(context.TODO(), q, &builds)

	return builds, err
}

func (g *GCPCloud) ReleaseList(app string, limit int) (pb.Releases, error) {
	q := datastore.NewQuery(releaseKind).
		Ancestor(g.stackKey()).
		Filter("App = ", app).
		Limit(limit)

	var rs pb.Releases
	_, err := g.datastore().GetAll(context.TODO(), q, &rs)

	return rs, err
}

func (g *GCPCloud) ReleaseDelete(app, id string) error {
	return g.datastore().Delete(context.TODO(), g.nestedKey(releaseKind, id))
}

func (g *GCPCloud) BuildImport(gskey string, tarf []byte) error {
	service := g.storage()
	bucket := g.BucketName

	log.Infof("Pushing code to gs://%s/%s", bucket, gskey)

	object := &storage.Object{
		Bucket:      bucket,
		Name:        gskey,
		ContentType: "application/gzip",
	}

	if _, err := service.Objects.Insert(bucket, object).Media(bytes.NewBuffer(tarf)).Do(); err != nil {
		return fmt.Errorf("Uploading to gs://%s/%s err: %s", bucket, gskey, err)
	}

	return nil
}

func (g *GCPCloud) BuildCreate(app string, tarf []byte) (*pb.Build, error) {
	service := g.storage()
	bucket := g.BucketName
	id := generateId("B", 5)
	gskey := fmt.Sprintf("%s.tar.gz", id)

	log.Infof("Pushing code to gs://%s/%s", bucket, gskey)

	object := &storage.Object{
		Bucket:      bucket,
		Name:        gskey,
		ContentType: "application/gzip",
	}

	if _, err := service.Objects.Insert(bucket, object).Media(bytes.NewBuffer(tarf)).Do(); err != nil {
		return nil, fmt.Errorf("Uploading to gs://%s/%s err: %s", bucket, gskey, err)
	}

	cb := g.cloudbuilder()

	log.Infof("Building from gs://%s/%s", bucket, gskey)
	tag := fmt.Sprintf("gcr.io/$PROJECT_ID/%v:%v", app, id)
	latestTag := fmt.Sprintf("gcr.io/$PROJECT_ID/%v:latest", app)

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

		return nil, fmt.Errorf("failed to initiate build %v", err)
	}

	remoteId, err := getBuildID(op)
	if err != nil {
		return nil, fmt.Errorf("failed to get Id for build %v", err)
	}

	build := &pb.Build{
		App:       app,
		Id:        id,
		RemoteId:  remoteId,
		Status:    pb.Status_CREATED,
		CreatedAt: timestampNow(),
	}

	log.Debugf("Saving build %s", toJson(build))

	if _, err := g.datastore().Put(context.TODO(), g.nestedKey(buildKind, build.Id), build); err != nil {
		return nil, fmt.Errorf("saving build err: %v", err)
	}

	return build, nil
}

func (g *GCPCloud) BuildLogs(app, id string, index int) (int, []string, error) {
	storageService := g.storage()
	i, logs, err := buildLogs(storageService, g.BucketName, id, index)
	if err != nil {
		return i, logs, err
	}

	return i, logs, err
}

func (g *GCPCloud) BuildRelease(b *pb.Build) (*pb.Release, error) {
	image := fmt.Sprintf("gcr.io/%v/%v:%v", g.Project, b.App, b.Id)
	log.Debugf("---- Docker Image: %s", image)

	envVars, err := g.EnvironmentGet(b.App)
	if err != nil {
		return nil, err
	}

	deployer, err := newDeployer(g.DeploymentName)
	if err != nil {
		return nil, err
	}

	port := 8080
	if pv, ok := envVars["PORT"]; ok {
		p, err := strconv.Atoi(pv)
		if err != nil {
			return nil, err
		}
		port = p
	}

	ret, err := deployer.Run(&DeployRequest{
		ServiceID:     b.App,
		Image:         image,
		Replicas:      1,
		Environment:   g.DeploymentName,
		Zone:          g.Zone,
		ContainerPort: intstr.FromInt(port),
		EnvVars:       envVars,
	})

	if err != nil {
		return nil, err
	}

	log.Debugf("Deploying %s with %s", b.App, toJson(ret.Request))

	r := &pb.Release{
		Id:        generateId("R", 5),
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.Status_CREATED,
		CreatedAt: timestampNow(),
	}

	_, err = g.datastore().Put(context.TODO(), g.nestedKey(releaseKind, r.Id), r)

	return r, err
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
	key := fmt.Sprintf("log-%s.txt", bid)
	resp, err := service.Objects.Get(bucket, key).Download()
	lines := []string{}

	if err != nil {
		return index, lines, err
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
