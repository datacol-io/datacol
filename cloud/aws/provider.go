package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AwsCloud struct {
	DeploymentName        string
	Region                string
	Access, Secret, Token string
}

func (p *AwsCloud) config() *aws.Config {
	config := &aws.Config{}

	if p.Region != "" {
		config.Region = aws.String(p.Region)
	}

	return config
}

func (p *AwsCloud) cloudformation() *cloudformation.CloudFormation {
	return cloudformation.New(session.New(), p.config())
}

func (p *AwsCloud) dynamodb() *dynamodb.DynamoDB {
	return dynamodb.New(session.New(), p.config())
}

func (p *AwsCloud) s3() *s3.S3 {
	return s3.New(session.New(), p.config())
}

func (p *AwsCloud) codebuild() *codebuild.CodeBuild {
	return codebuild.New(session.New(), p.config())
}
