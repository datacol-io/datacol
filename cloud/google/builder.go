package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud/common"
	sched "github.com/dinesh/datacol/cloud/kube"
	"google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	buildKind   = "Build"
	releaseKind = "Release"
)

func (g *GCPCloud) BuildGet(app, id string) (*pb.Build, error) {
	var b pb.Build
	ctx, key := g.nestedKey(buildKind, id)
	if err := g.datastore().Get(ctx, key, &b); err != nil {
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
			return &b, g.saveBuild(&b)
		}
	}

	return &b, nil
}

func (g *GCPCloud) BuildDelete(app, id string) error {
	ctx, key := g.nestedKey(buildKind, id)
	return g.datastore().Delete(ctx, key)
}

func (g *GCPCloud) BuildList(app string, limit int) (pb.Builds, error) {
	q := datastore.NewQuery(buildKind).Namespace(g.DeploymentName).Filter("app = ", app).Limit(limit)

	var builds pb.Builds
	_, err := g.datastore().GetAll(context.Background(), q, &builds)

	return builds, err
}

func (g *GCPCloud) ReleaseList(app string, limit int) (pb.Releases, error) {
	q := datastore.NewQuery(releaseKind).Namespace(g.DeploymentName).Filter("app = ", app).Limit(limit)

	var rs pb.Releases
	_, err := g.datastore().GetAll(context.Background(), q, &rs)

	return rs, err
}

func (g *GCPCloud) ReleaseDelete(app, id string) error {
	ctx, key := g.nestedKey(buildKind, id)
	return g.datastore().Delete(ctx, key)
}

func (g *GCPCloud) BuildCreate(app string, req *pb.CreateBuildOptions) (*pb.Build, error) {
	id := generateId("B", 5)

	build := &pb.Build{
		App:       app,
		Id:        id,
		Status:    "CREATED",
		Procfile:  req.Procfile,
		CreatedAt: timestampNow(),
	}

	return build, g.saveBuild(build)
}

func (g *GCPCloud) BuildImport(id, filename string) error {
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
				Name: "gcr.io/cloud-builders/docker:17.05",
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

	build.RemoteId = remoteId
	return g.saveBuild(build)
}

func (g *GCPCloud) saveBuild(b *pb.Build) error {
	log.Debugf("Saving build %s", toJson(b))

	ctx, key := g.nestedKey(buildKind, b.Id)
	if _, err := g.datastore().Put(ctx, key, b); err != nil {
		return fmt.Errorf("saving build err: %v", err)
	}

	return nil
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

func (g *GCPCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	image := fmt.Sprintf("gcr.io/%v/%v:%v", g.Project, b.App, b.Id)
	log.Debugf("---- Docker Image: %s", image)

	envVars, err := g.EnvironmentGet(b.App)
	if err != nil {
		return nil, err
	}

	app, err := g.AppGet(b.App)
	if err != nil {
		return nil, err
	}

	domains := sched.MergeAppDomains(app.Domains, options.Domain)

	deployer, err := sched.NewDeployer(g.kubeClient())
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

	command, proctype, err := common.GetContainerCommand(b)
	if err != nil {
		return nil, err
	}

	ret, err := deployer.Run(&sched.DeployRequest{
		ServiceID:     common.GetJobID(b.App, proctype),
		App:           b.App,
		Proctype:      proctype,
		Args:          command,
		Image:         image,
		Namespace:     g.DeploymentName,
		Zone:          g.DefaultZone,
		ContainerPort: intstr.FromInt(port),
		EnvVars:       envVars,
		Domains:       domains,
	})

	if err != nil {
		return nil, err
	}

	if len(app.Domains) != len(domains) {
		app.Domains = domains
		ctx, key := g.nestedKey(appKind, app.Name)
		if _, err = g.datastore().Put(ctx, key, app); err != nil {
			log.Warnf("datastore put failed: %v", err)
		}
	}

	log.Debugf("Deployed %s with %s", b.App, toJson(ret.Request))

	r := &pb.Release{
		Id:        generateId("R", 5),
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
	}

	ctx, key := g.nestedKey(releaseKind, r.Id)
	_, err = g.datastore().Put(ctx, key, r)

	if err != nil {
		return r, err
	}

	app.ReleaseId = r.Id
	app.BuildId = b.Id

	return r, g.saveApp(app)
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
