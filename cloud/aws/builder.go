package aws

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/common"
	"github.com/ejholmes/cloudwatch"
)

func (a *AwsCloud) ecrRepository(n string) string {
	return fmt.Sprintf("%s-%s-repo", a.DeploymentName, n)
}

func (a *AwsCloud) codebuildProjectName(n string) string {
	return fmt.Sprintf("%s-%s-code-builder", a.DeploymentName, n)
}

func (a *AwsCloud) codeBuildBucket() string {
	return a.SettingBucket
}

func (a *AwsCloud) BuildGet(app, id string) (*pb.Build, error) {
	build, err := a.store.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	if build.Status == "IN_PROGRESS" {
		rb, err := a.fetchRemoteBuild(build.RemoteId)
		if err != nil {
			return nil, err
		}
		status := rb.BuildStatus
		if status != nil && *status != build.Status {
			build.Status = *status
			a.store.BuildSave(build)
		}
	}

	return build, nil
}

func (a *AwsCloud) BuildDelete(app, id string) error {
	return notImplemented
}

func (a *AwsCloud) BuildList(app string, limit int64) (pb.Builds, error) {
	return a.store.BuildList(app, limit)
}

func (a *AwsCloud) BuildStatus(id string, status string) error {
	build, err := a.BuildGet("", id)
	if err != nil {
		return err
	}
	build.Status = status

	return a.store.BuildSave(build)
}

func (a *AwsCloud) BuildImport(id string, tr io.Reader, w io.WriteCloser) error {
	build, err := a.BuildGet("", id)
	if err != nil {
		return err
	}
	dkr, err := a.dockerClient()
	if err != nil {
		return fmt.Errorf("docker client: %v", err)
	}

	app, id := build.App, build.Id
	target := fmt.Sprintf("%s/%s", a.dockerRegistryURL(), app)

	err = common.BuildDockerLoad(target, id, dkr, tr, w)
	if err == nil {
		build.Status = "SUCCEEDED"
	} else {
		build.Status = "FAILED"
	}

	// Ignore the error by build save
	if berr := a.store.BuildSave(build); berr != nil {
		log.Errorf("Failed to save build: %v", berr)
	}

	return err
}

func (a *AwsCloud) BuildUpload(id, gzipPath string) error {
	build, err := a.BuildGet("", id)
	if err != nil {
		return err
	}
	app := build.App

	// We need to convert gzip to zip format since AWS codebuild only
	// supports zip file for s3 based source
	log.Debugf("converting gzip to zip of %s", gzipPath)
	zipPath, err := convertGzipToZip(app, gzipPath)
	if err != nil {
		return fmt.Errorf("converting gzip to zip archive. err: %v", err)
	}

	defer os.RemoveAll(zipPath)

	reader, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("reading tempfile err: %v", err)
	}
	defer reader.Close()
	log.Debugf("Uploading to s3 from %s", zipPath)

	uploader := s3manager.NewUploaderWithClient(a.s3())

	if _, err = uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: aws.String(a.codeBuildBucket()),
		Key:    aws.String(app + "/source.zip"),
	}, func(u *s3manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024 // 64MB per part
	}); err != nil {
		return fmt.Errorf("uploading source to s3 err: %v", err)
	}

	log.Debug("OK \n")

	return a.startBuild(build, &pb.CreateBuildOptions{})
}

func (a *AwsCloud) BuildCreate(app string, req *pb.CreateBuildOptions) (*pb.Build, error) {
	Id := generateId("B", 5)

	build := &pb.Build{
		App:       app,
		Id:        Id,
		Version:   req.Version,
		Procfile:  req.Procfile,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
	}

	//FIXME: If version is not blank, we can trigger the build. this is a hack as of now and
	// should be replaced with better build API
	// Version os GIT COMMIT hash
	if req.Version != "" {
		return build, a.startBuild(build, req)
	}

	return build, a.store.BuildSave(build)
}

func (a *AwsCloud) BuildLogs(app, id string, index int) (int, []string, error) {
	rb, err := a.fetchRemoteBuild(id)
	if err != nil {
		return 0, nil, err
	}

	return a.buildLogs(a.s3(), rb.Logs.GroupName, rb.Logs.StreamName, id, index)
}

func (a *AwsCloud) BuildLogsStream(id string) (io.Reader, error) {
	fmt.Print("waiting for cloudwatch logstream (1s)")
	var rb *codebuild.Build

	bb, err := a.BuildGet("", id)
	if err != nil {
		return nil, err
	}

	if bb.RemoteId == "" {
		return nil, fmt.Errorf("No build process found for Id=%s", bb.Id)
	}

	for {
		b, err := a.fetchRemoteBuild(bb.RemoteId)
		if err != nil {
			return nil, fmt.Errorf("fetching build err: %v", err)
		}
		rb = b
		if rb.Logs != nil && rb.Logs.StreamName != nil {
			break
		}

		fmt.Print(".")
		time.Sleep(1 * time.Second)
	}

	fmt.Print(" OK\n")
	log.Debugf("Will start streaming from stream: %s", *rb.Logs.StreamName)

	return cloudwatch.NewGroup(*rb.Logs.GroupName, a.cloudwatchlogs()).Open(*rb.Logs.StreamName)
}

func (a *AwsCloud) fetchRemoteBuild(id string) (*codebuild.Build, error) {
	out, err := a.codebuild().BatchGetBuilds(&codebuild.BatchGetBuildsInput{
		Ids: []*string{aws.String(id)},
	})

	if err != nil {
		return nil, fmt.Errorf("fetching cloud builds err: %v", err)
	}
	for _, b := range out.Builds {
		if b.Id != nil && *b.Id == id {
			return b, nil
		}
	}

	return nil, fmt.Errorf("no build found by id: %s", id)
}

func (a *AwsCloud) buildLogs(svc *s3.S3, bucket, key *string, bid string, index int) (int, []string, error) {
	out, err := svc.GetObject(&s3.GetObjectInput{Bucket: bucket, Key: key})
	if err != nil {
		return 0, nil, err
	}

	defer out.Body.Close()
	lines := []string{}

	body, err := ioutil.ReadAll(out.Body)
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

func (a *AwsCloud) startBuild(build *pb.Build, req *pb.CreateBuildOptions) error {
	log.Infof("Starting the build ...")

	ret, err := a.codebuild().StartBuild(&codebuild.StartBuildInput{
		ProjectName:   aws.String(a.codebuildProjectName(build.App)),
		SourceVersion: aws.String(req.Version),
		EnvironmentVariablesOverride: []*codebuild.EnvironmentVariable{
			{
				Name:  aws.String("APP"),
				Value: aws.String(build.App),
			}, {
				Name:  aws.String("IMAGE_TAG"),
				Value: aws.String(build.Id),
			},
		},
	})

	if err != nil {
		return err
	}

	build.RemoteId = *ret.Build.Id
	build.Status = *ret.Build.BuildStatus

	log.Debugf("Persisting build: %+v", build)

	return a.store.BuildSave(build)
}
