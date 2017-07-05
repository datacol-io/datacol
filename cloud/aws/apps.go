package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/dinesh/datacol/api/models"
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
		return nil, err
	}

	if res.Item == nil {
		return nil, fmt.Errorf("no such app found by name: %s", name)
	}

	app := a.appFromItem(res.Item)
	return app, nil
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
