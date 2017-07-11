package aws

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
)

func (a *AwsCloud) dynamoApps() string {
	return fmt.Sprintf("%s-apps", a.DeploymentName)
}

func (a *AwsCloud) appFromItem(item map[string]*dynamodb.AttributeValue) *pb.App {
	name := coalesce(item["name"], "")

	return &pb.App{
		Name:      name,
		Status:    coalesce(item["status"], ""),
		ReleaseId: coalesce(item["release_id"], ""),
		Endpoint:  coalesce(item["endpoint"], ""),
	}
}

func (a *AwsCloud) AppList() (pb.Apps, error) {
	req := &dynamodb.QueryInput{
		ScanIndexForward: aws.Bool(false),
		TableName:        aws.String(a.dynamoApps()),
	}

	res, err := a.dynamodb().Query(req)
	if err != nil {
		return nil, err
	}

	apps := make(pb.Apps, len(res.Items))

	for i, item := range res.Items {
		apps[i] = a.appFromItem(item)
	}

	return apps, nil
}

func (a *AwsCloud) AppCreate(name string) (*pb.App, error) {
	app := &pb.App{Name: name, Status: pb.StatusCreated}
	req := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"name":   {S: aws.String(app.Name)},
			"status": {S: aws.String(app.Status)},
		},
		TableName: aws.String(a.dynamoApps()),
	}

	_, err := a.dynamodb().PutItem(req)
	return app, err
}

func (a *AwsCloud) AppRestart(app string) error {
	return nil
}

func (a *AwsCloud) AppGet(name string) (*pb.App, error) {
	req := &dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"name": {S: aws.String(name)},
		},
		TableName: aws.String(a.dynamoApps()),
	}

	res, err := a.dynamodb().GetItem(req)
	if err != nil {
		return nil, fmt.Errorf("fetching from dynamodb err: %v", err)
	}

	if res.Item == nil {
		return nil, fmt.Errorf("no such app found by name: %s", name)
	}

	app := a.appFromItem(res.Item)

	kc, err := getKubeClientset(a.DeploymentName)
	if err != nil {
		log.Warn(err)
		return app, nil
	}

	if app.Endpoint, err = sched.GetServiceEndpoint(kc, a.DeploymentName, name); err != nil {
		return app, err
	}

	return app, a.saveApp(app)
}

func (a *AwsCloud) AppDelete(name string) error {
	_, err := a.dynamodb().DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {S: aws.String(name)},
		},
		TableName: aws.String(a.dynamoApps()),
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *AwsCloud) saveApp(a *pb.App) error {
	req := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"name": {S: aws.String(a.Name)},
		},
		TableName: aws.String(p.dynamoApps()),
	}

	if a.Status != "" {
		req.Item["status"] = &dynamodb.AttributeValue{S: aws.String(a.Status)}
	}

	if a.Endpoint != "" {
		req.Item["endpoint"] = &dynamodb.AttributeValue{S: aws.String(a.Endpoint)}
	}

	if a.ReleaseId != "" {
		req.Item["release_id"] = &dynamodb.AttributeValue{S: aws.String(a.ReleaseId)}
	}

	_, err := p.dynamodb().PutItem(req)
	return err
}
