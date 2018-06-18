package store

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/datacol-io/datacol/api/models"
)

func (a *DynamoDBStore) ReleaseList(app string, limit int64) (pb.Releases, error) {
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

func (a *DynamoDBStore) ReleaseSave(r *pb.Release) error {
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

func (a *DynamoDBStore) ReleaseDelete(app, id string) error {
	return nil
}

func (a *DynamoDBStore) ReleaseCount(app string) (version int64) {
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

func (a *DynamoDBStore) releaseFromItem(item map[string]*dynamodb.AttributeValue) *pb.Release {
	return &pb.Release{
		Id:        coalesce(item["id"], ""),
		App:       coalesce(item["app"], ""),
		BuildId:   coalesce(item["build_id"], ""),
		Status:    coalesce(item["status"], ""),
		CreatedAt: int64(coalesceInt(item["created_at"], 0)),
		Version:   int64(coalesceInt(item["version"], 0)),
	}
}
