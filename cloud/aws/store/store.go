package store

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/datacol-io/datacol/api/store"
)

type DynamoDBStore struct {
	DeploymentName string
	SettingBucket  string

	*dynamodb.DynamoDB
}

var _ store.Store = &DynamoDBStore{}

func (p *DynamoDBStore) dynamodb() *dynamodb.DynamoDB {
	return p.DynamoDB
}

func (a *DynamoDBStore) dynamoApps() string {
	return fmt.Sprintf("%s-apps", a.DeploymentName)
}

func (a *DynamoDBStore) dynamoBuilds() string {
	return fmt.Sprintf("%s-builds", a.DeploymentName)
}

func (a *DynamoDBStore) dynamoReleases() string {
	return fmt.Sprintf("%s-releases", a.DeploymentName)
}
