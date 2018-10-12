package store

import (
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/datacol-io/datacol/api/models"
)

var (
	notImplemented = errors.New("Not Implemented yet.")
)

func (a *DynamoDBStore) BuildCreate(app string, req *pb.CreateBuildOptions) (*pb.Build, error) {
	return nil, notImplemented
}

func (a *DynamoDBStore) BuildGet(app, id string) (*pb.Build, error) {
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

	return build, nil
}

func (a *DynamoDBStore) BuildDelete(app, id string) error {
	return notImplemented
}

func (a *DynamoDBStore) BuildList(app string, limit int64) (pb.Builds, error) {
	req := &dynamodb.ScanInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(a.dynamoBuilds()),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":app": {S: aws.String(app)},
		},
		FilterExpression: aws.String("app=:app"),
	}

	res, err := a.dynamodb().Scan(req)
	if err != nil {
		return nil, err
	}

	builds := make(pb.Builds, len(res.Items))

	for i, item := range res.Items {
		builds[i] = a.buildFromItem(item)
	}

	sort.Slice(builds, func(i, j int) bool {
		return builds[i].CreatedAt > builds[j].CreatedAt
	})

	size := int64(len(builds))

	if limit > 0 && size < limit {
		limit = size
	}

	return builds[0:limit], nil
}

func (a *DynamoDBStore) buildFromItem(item map[string]*dynamodb.AttributeValue) *pb.Build {
	return &pb.Build{
		Id:        coalesce(item["id"], ""),
		App:       coalesce(item["app"], ""),
		Status:    coalesce(item["status"], ""),
		Version:   coalesce(item["version"], ""),
		RemoteId:  coalesce(item["remote_id"], ""),
		Procfile:  coalesceBytes(item["procfile"]),
		CreatedAt: int64(coalesceInt(item["created_at"], 0)),
	}
}

func (a *DynamoDBStore) BuildSave(b *pb.Build) error {
	if b.Id == "" {
		b.Id = generateId("B", 5)
	}

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

	if b.Version != "" {
		req.Item["version"] = &dynamodb.AttributeValue{S: aws.String(b.Version)}
	}

	if b.RemoteId != "" {
		req.Item["remote_id"] = &dynamodb.AttributeValue{S: aws.String(b.RemoteId)}
	}

	if len(b.Procfile) > 0 {
		req.Item["procfile"] = &dynamodb.AttributeValue{B: b.Procfile}
	}

	if b.CreatedAt > 0 {
		req.Item["created_at"] = &dynamodb.AttributeValue{
			N: aws.String(fmt.Sprintf("%d", b.CreatedAt)),
		}
	}

	_, err := a.dynamodb().PutItem(req)
	return err
}
