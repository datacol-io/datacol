package aws

import (
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/cloudformation"
)

type AwsCloud struct {
  DeploymentName string
  Region string
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
