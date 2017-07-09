package aws

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
	"github.com/ejholmes/cloudwatch"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func (a *AwsCloud) ecrRepository() string {
	return fmt.Sprintf("%s-repo", a.DeploymentName)
}

func (a *AwsCloud) codebuildProjectName() string {
	return fmt.Sprintf("%s-code-builder", a.DeploymentName)
}

func (a *AwsCloud) dynamoBuilds() string {
	return fmt.Sprintf("%s-builds", a.DeploymentName)
}

func (a *AwsCloud) codeBuildBucket() string {
	return a.SettingBucket
}

func (g *AwsCloud) GetRunningPods(app string) (string, error) {
	ns := g.DeploymentName
	c, err := getKubeClientset(ns)
	if err != nil {
		return "", err
	}

	return sched.RunningPods(ns, app, c)
}

func (a *AwsCloud) BuildGet(app, id string) (*pb.Build, error) {
	req := &dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
		TableName: aws.String(a.dynamoBuilds()),
	}

	res, err := a.dynamodb().GetItem(req)
	if err != nil {
		return nil, err
	}

	if res.Item == nil {
		return nil, fmt.Errorf("no build found by id: %s", id)
	}

	build := a.buildFromItem(res.Item)

	if build.Status == "IN_PROGRESS" {
		rb, err := a.fetchRemoteBuild(build.RemoteId)
		if err != nil {
			return nil, err
		}
		status := rb.BuildStatus
		if status != nil && *status != build.Status {
			build.Status = *status
			a.buildSave(build)
		}
	}

	return build, nil
}

func (a *AwsCloud) BuildDelete(app, id string) error {
	return nil
}

func (a *AwsCloud) BuildList(app string, limit int) (pb.Builds, error) {
	return nil, nil
}

func (a *AwsCloud) BuildCreate(app, gzipPath string) (*pb.Build, error) {
	zipPath, err := convertGzipToZip(gzipPath)
	if err != nil {
		return nil, fmt.Errorf("converting gzip to zip archive. err: %v", err)
	}
	defer os.RemoveAll(zipPath)

	reader, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("reading tempfile err: %v", err)
	}
	defer reader.Close()
	fileInfo, _ := reader.Stat()

	buffer := make([]byte, fileInfo.Size())
	reader.Read(buffer)

	fileBytes := bytes.NewReader(buffer)

	if _, err := a.s3().PutObject(&s3.PutObjectInput{
		Body:   fileBytes,
		Bucket: aws.String(a.codeBuildBucket()),
		Key:    aws.String("source.zip"),
	}); err != nil {
		return nil, fmt.Errorf("uploading source to s3 err: %v", err)
	}

	build := &pb.Build{
		App:       app,
		Id:        generateId("B", 5),
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
	}

	if err := a.buildSave(build); err != nil {
		return nil, fmt.Errorf("saving to dynamodb err: %v", err)
	}

	log.Infof("Starting the build ...")
	ret, err := a.codebuild().StartBuild(&codebuild.StartBuildInput{
		ProjectName: aws.String(a.codebuildProjectName()),
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
		return nil, err
	}

	build.RemoteId = *ret.Build.Id
	build.Status = *ret.Build.BuildStatus

	log.Debugf("persisting build: %+v", build)
	if err = a.buildSave(build); err != nil {
		return nil, err
	}

	return build, nil
}

func (a *AwsCloud) BuildImport(key string, tarf []byte) error {
	return nil
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

	for {
		b, err := a.fetchRemoteBuild(id)
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

func (a *AwsCloud) buildFromItem(item map[string]*dynamodb.AttributeValue) *pb.Build {
	return &pb.Build{
		Id:       coalesce(item["id"], ""),
		App:      coalesce(item["app"], ""),
		Status:   coalesce(item["status"], ""),
		RemoteId: coalesce(item["remote_id"], ""),
	}
}

func (a *AwsCloud) buildSave(b *pb.Build) error {
	req := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id":  {S: aws.String(b.Id)},
			"app": {S: aws.String(b.App)},
		},
		TableName: aws.String(a.dynamoBuilds()),
	}

	if b.Status != "" {
		req.Item["status"] = &dynamodb.AttributeValue{S: aws.String(b.Status)}
	}

	if b.RemoteId != "" {
		req.Item["remote_id"] = &dynamodb.AttributeValue{S: aws.String(b.RemoteId)}
	}

	_, err := a.dynamodb().PutItem(req)
	return err
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
