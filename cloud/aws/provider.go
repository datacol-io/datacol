package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AwsCloud struct {
	DeploymentName        string
	Region, SettingBucket string
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

func (p *AwsCloud) cloudwatchlogs() *cloudwatchlogs.CloudWatchLogs {
	return cloudwatchlogs.New(session.New(), p.config())
}

func (p *AwsCloud) describeStack() (*cloudformation.Stack, error) {
	out, err := p.cloudformation().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(p.DeploymentName),
	})

	if err != nil {
		return nil, err
	}

	for _, s := range out.Stacks {
		if s.StackName != nil && *s.StackName == p.DeploymentName {
			return s, nil
		}
	}

	return nil, fmt.Errorf("No stack found by name: %s", p.DeploymentName)
}
