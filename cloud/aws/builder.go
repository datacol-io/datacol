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
	"github.com/ejholmes/cloudwatch"
	"io"
	"io/ioutil"
	"strings"
)

func (a *AwsCloud) dynamoBuilds() string {
	return fmt.Sprintf("%s-builds", a.DeploymentName)
}

func (a *AwsCloud) codeBuildBucket() string {
	return a.SettingBucket
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

func (a *AwsCloud) BuildCreate(app string, tarf []byte) (*pb.Build, error) {
	if _, err := a.s3().PutObject(&s3.PutObjectInput{
		Body:        bytes.NewReader(tarf),
		Bucket:      aws.String(a.codeBuildBucket()),
		Key:         aws.String("source.zip"),
		ContentType: aws.String("application/zip"),
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
		EnvironmentVariablesOverride: []*codebuild.EnvironmentVariable{
			{
				Name:  aws.String("APP"),
				Value: aws.String(build.App),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	build.RemoteId = *ret.Build.Id
	build.Status = *ret.Build.BuildStatus

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
	rb, err := a.fetchRemoteBuild(id)
	if err != nil {
		return nil, err
	}

	return cloudwatch.NewGroup(*rb.Logs.GroupName, a.cloudwatchlogs()).Open(*rb.Logs.StreamName)
}

func (a *AwsCloud) BuildRelease(b *pb.Build) (*pb.Release, error) {
	return nil, nil
}

func (a *AwsCloud) buildFromItem(item map[string]*dynamodb.AttributeValue) *pb.Build {
	return &pb.Build{
		Id:     coalesce(item["id"], ""),
		App:    coalesce(item["app"], ""),
		Status: coalesce(item["status"], ""),
	}
}

func (a *AwsCloud) buildSave(b *pb.Build) error {
	req := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id":        {S: aws.String(b.Id)},
			"app":       {S: aws.String(b.App)},
			"status":    {S: aws.String(b.Status)},
			"remote_id": {S: aws.String(b.RemoteId)},
		},
		TableName: aws.String(a.dynamoBuilds()),
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
