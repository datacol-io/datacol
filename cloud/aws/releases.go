package aws

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
	kube "github.com/datacol-io/datacol/k8s"
)

func (a *AwsCloud) dynamoReleases() string {
	return fmt.Sprintf("%s-releases", a.DeploymentName)
}

func (a *AwsCloud) ReleaseList(app string, limit int) (pb.Releases, error) {
	req := &dynamodb.ScanInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(a.dynamoReleases()),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":app": {S: aws.String(app)},
		},
		FilterExpression: aws.String("app=:app"),
	}

	res, err := a.dynamodb().Scan(req)
	if err != nil {
		return nil, err
	}

	releases := make(pb.Releases, len(res.Items))

	for i, item := range res.Items {
		releases[i] = a.releaseFromItem(item)
	}

	return releases, nil
}

func (a *AwsCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	image := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		os.Getenv("AWS_ACCOUNT_ID"), a.Region, a.ecrRepository(b.App), b.Id,
	)
	log.Debugf("---- Docker Image: %s", image)

	app, err := a.AppGet(b.App)
	if err != nil {
		return nil, err
	}

	envVars, err := a.EnvironmentGet(b.App)
	if err != nil {
		return nil, err
	}

	r := &pb.Release{
		Id:        generateId("R", 5),
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
		Version:   a.releaseCount(b.App) + 1,
	}

	if err = a.releaseSave(r); err != nil {
		return r, err
	}

	domains := app.Domains
	for _, domain := range strings.Split(options.Domain, ",") {
		domains = kube.MergeAppDomains(domains, domain)
	}

	if len(app.Domains) != len(domains) {
		app.Domains = domains
	}

	app.BuildId = b.Id
	app.ReleaseId = r.Id
	rversion := fmt.Sprintf("%d", r.Version)

	log.Debugf("Saving app state: %s err:%v", toJson(app), a.saveApp(app)) // note the mutate function

	if err := common.UpdateApp(a.kubeClient(), b, a.DeploymentName,
		image, false, domains, envVars, cloud.AwsProvider, rversion); err != nil {
		return nil, err
	}

	//TODO: update release status

	return r, nil
}

func (a *AwsCloud) releaseSave(r *pb.Release) error {
	req := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id":         {S: aws.String(r.Id)},
			"app":        {S: aws.String(r.App)},
			"build_id":   {S: aws.String(r.BuildId)},
			"created_at": {N: aws.String(fmt.Sprintf("%d", r.CreatedAt))},
		},
		TableName: aws.String(a.dynamoReleases()),
	}

	if r.Status != "" {
		req.Item["status"] = &dynamodb.AttributeValue{S: aws.String(r.Status)}
	}

	if r.Version > 0 {
		req.Item["version"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", r.Version))}
	}

	_, err := a.dynamodb().PutItem(req)
	return err
}

func (a *AwsCloud) releaseFromItem(item map[string]*dynamodb.AttributeValue) *pb.Release {
	return &pb.Release{
		Id:        coalesce(item["id"], ""),
		App:       coalesce(item["app"], ""),
		BuildId:   coalesce(item["build_id"], ""),
		Status:    coalesce(item["status"], ""),
		CreatedAt: int32(coalesceInt(item["created_at"], 0)),
		Version:   int64(coalesceInt(item["version"], 0)),
	}
}

func (a *AwsCloud) ReleaseDelete(app, id string) error {
	return notImplemented
}

func (a *AwsCloud) releaseCount(app string) (version int64) {
	a.lock.Lock()
	defer a.lock.Unlock()

	queryInput := &dynamodb.ScanInput{
		TableName: aws.String(a.dynamoReleases()),
		Select:    aws.String("COUNT"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":app": {S: aws.String(app)},
		},
		FilterExpression: aws.String("app=:app"),
	}

	res, err := a.dynamodb().Scan(queryInput)
	if err != nil {
		log.Warnf("Fetching release count: %v", err)
	} else {
		version = *res.ScannedCount
	}

	return version
}
