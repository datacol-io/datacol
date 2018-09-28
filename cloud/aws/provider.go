package aws

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/datacol-io/datacol/api/store"
	dynamo_store "github.com/datacol-io/datacol/cloud/aws/store"
	docker "github.com/fsouza/go-dockerclient"
)

type AwsCloud struct {
	DeploymentName        string
	Region, SettingBucket string
	Access, Secret, Token string

	lock  sync.Mutex
	store store.Store
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

func (p *AwsCloud) ecr() *ecr.ECR {
	return ecr.New(session.New(), p.config())
}

func (p *AwsCloud) dockerRegistryURL() string {
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", os.Getenv("AWS_ACCOUNT_ID"), os.Getenv("AWS_REGION"))
}

func (p *AwsCloud) dockerClient() (*docker.Client, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	return client, err
}

func (p *AwsCloud) dockerLogin() error {
	tres, err := p.ecr().GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		log.Printf("ecr auth token: %v\n", err)
		return err
	}

	if len(tres.AuthorizationData) != 1 {
		log.Println("no authorization data")
		return fmt.Errorf("no authorization data")
	}

	auth, err := base64.StdEncoding.DecodeString(*tres.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		log.Println("encode token", err)
		return err
	}

	authParts := strings.SplitN(string(auth), ":", 2)
	if len(authParts) != 2 {
		log.Println("invalid auth data")
		return fmt.Errorf("invalid auth data")
	}

	registry, err := url.Parse(*tres.AuthorizationData[0].ProxyEndpoint)
	if err != nil {
		return err
	}

	out, err := exec.Command("docker", "login", "-u", authParts[0], "-p", authParts[1], registry.Host).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s\n", lastline(out), err.Error())
	}

	return nil
}

func (p *AwsCloud) describeStack(name string) (*cloudformation.Stack, error) {
	if name == "" {
		name = p.DeploymentName
	}

	out, err := p.cloudformation().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})

	if err != nil {
		return nil, err
	}

	for _, s := range out.Stacks {
		if s.StackName != nil && *s.StackName == name {
			return s, nil
		}
	}

	return nil, fmt.Errorf("No stack found by name: %s", name)
}

func (p *AwsCloud) describeStacks(input *cloudformation.DescribeStacksInput) ([]*cloudformation.Stack, error) {
	out, err := p.cloudformation().DescribeStacks(input)
	if err != nil {
		return nil, err
	}

	return out.Stacks, nil
}

func (p *AwsCloud) describeStackEvents(input *cloudformation.DescribeStackEventsInput) (*cloudformation.DescribeStackEventsOutput, error) {
	res, err := p.cloudformation().DescribeStackEvents(input)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *AwsCloud) Setup() {
	store := dynamo_store.DynamoDBStore{
		DeploymentName: p.DeploymentName,
		SettingBucket:  p.SettingBucket,
		DynamoDB:       p.dynamodb(),
	}

	p.store = &store
}
