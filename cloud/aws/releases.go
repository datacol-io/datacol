package aws

import (
	"fmt"
	"os"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
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

	sort.Slice(releases, func(i, j int) bool {
		return releases[i].CreatedAt > releases[j].CreatedAt
	})

	if len(res.Items) < limit {
		limit = len(res.Items)
	}

	return releases[0:limit], nil
}

func (a *AwsCloud) ReleaseGet(app, id string) (*pb.Release, error) {
	req := &dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(a.dynamoReleases()),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
	}

	res, err := a.dynamodb().GetItem(req)
	if err != nil {
		return nil, err
	}

	if res.Item == nil {
		return nil, fmt.Errorf("no release found by id: %s", id)
	}

	release := a.releaseFromItem(res.Item)

	return release, nil
}

func (a *AwsCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	r := &pb.Release{
		Id:        generateId("R", 5),
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
		Version:   a.releaseCount(b.App) + 1,
	}

	if err := a.releaseSave(r); err != nil {
		return r, err
	}

	return r, nil
}

func (a *AwsCloud) ReleasePromote(name, id string) error {
	r, err := a.ReleaseGet(name, id)
	if err != nil {
		return err
	}

	if r.BuildId == "" {
		return fmt.Errorf("No build found for release: %s", id)
	}

	b, err := a.BuildGet(name, r.BuildId)
	if err != nil {
		return err
	}

	image := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		os.Getenv("AWS_ACCOUNT_ID"), a.Region, a.ecrRepository(name), r.BuildId,
	)

	app, err := a.AppGet(name)
	if err != nil {
		return err
	}

	envVars, err := a.EnvironmentGet(name)
	if err != nil {
		return err
	}

	rversion := fmt.Sprintf("%d", r.Version)

	if err = common.UpdateApp(a.kubeClient(), b, a.DeploymentName,
		image, false, app.Domains, envVars, cloud.AwsProvider, rversion); err != nil {
		return err
	}

	app.BuildId = b.Id
	app.ReleaseId = r.Id

	return a.saveApp(app)

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
