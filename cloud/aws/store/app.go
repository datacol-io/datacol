package store

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/datacol-io/datacol/api/models"
)

func (a *DynamoDBStore) AppList() (pb.Apps, error) {
	req := &dynamodb.ScanInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(a.dynamoApps()),
	}

	res, err := a.dynamodb().Scan(req)
	if err != nil {
		return nil, err
	}

	apps := make(pb.Apps, len(res.Items))

	for i, item := range res.Items {
		apps[i] = a.appFromItem(item)
	}

	return apps, nil
}

func (a *DynamoDBStore) AppCreate(app *pb.App, req *pb.AppCreateOptions) error {
	name := app.Name
	if _, err := a.AppGet(name); err == nil {
		return fmt.Errorf("Duplicate app: %s", name)
	}

	return a.AppUpdate(app)
}

func (a *DynamoDBStore) AppGet(name string) (*pb.App, error) {
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

	return app, nil
}

func (a *DynamoDBStore) AppDelete(name string) error {
	_, err := a.dynamodb().DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {S: aws.String(name)},
		},
		TableName: aws.String(a.dynamoApps()),
	})

	return err
}

func (p *DynamoDBStore) AppUpdate(a *pb.App) error {
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

	if a.BuildId != "" {
		req.Item["build_id"] = &dynamodb.AttributeValue{S: aws.String(a.BuildId)}
	}

	if a.ReleaseId != "" {
		req.Item["release_id"] = &dynamodb.AttributeValue{S: aws.String(a.ReleaseId)}
	}

	if len(a.Domains) > 0 {
		list := []*dynamodb.AttributeValue{}
		for _, d := range a.Domains {
			list = append(list, &dynamodb.AttributeValue{S: aws.String(d)})
		}

		req.Item["domains"] = &dynamodb.AttributeValue{L: list}
	}

	_, err := p.dynamodb().PutItem(req)
	return err
}

func (a *DynamoDBStore) appFromItem(item map[string]*dynamodb.AttributeValue) *pb.App {
	name := coalesce(item["name"], "")

	app := &pb.App{
		Name:      name,
		Status:    coalesce(item["status"], ""),
		ReleaseId: coalesce(item["release_id"], ""),
		BuildId:   coalesce(item["build_id"], ""),
		Endpoint:  coalesce(item["endpoint"], ""),
	}

	if domainValues, ok := item["domains"]; ok {
		domains := []string{}
		for _, key := range domainValues.L {
			domains = append(domains, coalesce(key, ""))
		}
		app.Domains = domains
	}

	return app
}
