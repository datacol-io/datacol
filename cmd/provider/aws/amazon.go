package aws

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/io"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"os"
	"strings"
	"time"
)

const (
	welcomeMessage = `Welcome to Datacol CLI. This command will guide you through creating a new infrastructure inside your AWS account. 
It uses various AWS services (like EC2, elastic container registry, Cloudformation, Cloudwatch etc) under the hood to 
automate all away to give you a better deployment experience.

Datacol CLI will authenticate with your AWS Account and install the Datacol platform into your AWS account. 
These credentials will only be used to communicate between this installer running on your computer and the AWS.`
)

var (
	bastionType     = "t2.nano"
	bucketPrefix    = "datacol"
	networkProvider = "calico"
	adminIngressLoc = "0.0.0.0/0"
)

type InitOptions struct {
	Name, Zone, Region, Bucket, KeyName string
	ApiKey, Version, ArtifactBucket     string
	ClusterVersion, MachineType         string
	DiskSize, NumNodes, ControllerPort  int
	UseSpotInstance                     bool
}

type initResponse struct {
	Host, Password string
}

type AwsCredentials struct {
	Access     string `json:"AccessKeyId"`
	Secret     string `json:"SecretAccessKey"`
	Session    string `json:"SessionToken"`
	Expiration time.Time
}

func InitializeStack(opts *InitOptions, creds *AwsCredentials) (*initResponse, error) {
	fmt.Printf(welcomeMessage)
	prompt("")

	cf := cloudformation.New(session.New(), awsConfig(opts.Region, creds))
	tmpl, err := io.ReadFile("./cmd/provider/aws/formation.yaml")

	if err != nil {
		return nil, err
	}

	log.Debugf("Creating stack with %+v", opts)
	req := &cloudformation.CreateStackInput{
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		Parameters: []*cloudformation.Parameter{
			{ParameterKey: aws.String("AvailabilityZone"), ParameterValue: aws.String(opts.Zone)},
			{ParameterKey: aws.String("KeyName"), ParameterValue: aws.String(opts.KeyName)},
			// {ParameterKey: aws.String("DiskSizeGb"), ParameterValue: aws.String(fmt.Sprintf("%d", opts.DiskSize))},
			{ParameterKey: aws.String("BastionInstanceType"), ParameterValue: aws.String(bastionType)},
			{ParameterKey: aws.String("AdminIngressLocation"), ParameterValue: aws.String(adminIngressLoc)},
			{ParameterKey: aws.String("NetworkingProvider"), ParameterValue: aws.String(networkProvider)},
			{ParameterKey: aws.String("K8sNodeCapacity"), ParameterValue: aws.String(fmt.Sprintf("%d", opts.NumNodes-1))},
			{ParameterKey: aws.String("InstanceType"), ParameterValue: aws.String(opts.MachineType)},
		},
		StackName:    aws.String(opts.Name),
		TemplateBody: aws.String(tmpl),
	}

	fmt.Printf("Creating a new stack: %s \n", opts.Name)

	res, err := cf.CreateStack(req)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "AlreadyExistsException" {
				return nil, fmt.Errorf("Stack %q already exists. Run `datacol destroy` then try again", opts.Name)
			}
		}
		return nil, fmt.Errorf("creating stack err: %v", err)
	}

	host, err := waitForCompletion(*res.StackId, cf, false)
	if err != nil {
		return nil, err
	}

	return &initResponse{Host: host, Password: opts.ApiKey}, nil
}

func TeardownStack(stack, region string, creds *AwsCredentials) error {
	cf := cloudformation.New(session.New(), awsConfig(region, creds))
	fmt.Println("Deleting stack", stack, " ...")
	return destroyRack(stack, cf)
}

func destroyRack(rackName string, cf *cloudformation.CloudFormation) error {
	if len(rackName) == 0 {
		return nil
	}

	for {
		var s *cloudformation.Stack
		res, err := cf.DescribeStacks(&cloudformation.DescribeStacksInput{})
		if err != nil {
			return err
		}

		for _, st := range res.Stacks {
			if st.StackName == nil || st.StackStatus == nil {
				continue
			}

			if rackName == *st.StackName {
				s = st
				break
			}
		}

		if s == nil {
			return nil
		}

		stackName := *s.StackName

		switch *s.StackStatus {
		case "CREATE_COMPLETE", "ROLLBACK_COMPLETE", "UPDATE_COMPLETE", "UPDATE_ROLLBACK_COMPLETE":
			deleteStack(stackName, cf)
		case "CREATE_FAILED", "DELETE_FAILED", "ROLLBACK_FAILED", "UPDATE_ROLLBACK_FAILED":
			eres, err := cf.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
				StackName: aws.String(stackName),
			})
			if err != nil {
				return err
			}

			for _, event := range eres.StackEvents {
				if strings.HasSuffix(*event.ResourceStatus, "FAILED") {
					fmt.Printf("Failed: %s: %s\n", *event.LogicalResourceId, *event.ResourceStatusReason)
				}
			}

			deleteStack(stackName, cf)
		case "DELETE_IN_PROGRESS":
			displayProgress(stackName, cf, true)
		default:
			displayProgress(stackName, cf, true)
		}

		time.Sleep(5 * time.Second)
	}
	return nil
}

func deleteStack(stack string, cf *cloudformation.CloudFormation) error {
	_, err := cf.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stack),
	})

	return err
}

func waitForCompletion(stack string, cf *cloudformation.CloudFormation, isDeleting bool) (string, error) {
	for {
		dres, err := cf.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stack),
		})
		if err != nil {
			return "", err
		}

		err = displayProgress(stack, cf, isDeleting)
		if err != nil {
			return "", err
		}

		if len(dres.Stacks) != 1 {
			return "", fmt.Errorf("could not read stack status")
		}

		switch *dres.Stacks[0].StackStatus {
		case "CREATE_COMPLETE":
			for _, o := range dres.Stacks[0].Outputs {
				if *o.OutputKey == "BastionHostPublicDNS" {
					return *o.OutputValue, nil
				}
			}

			return "", fmt.Errorf("could not install stack")
		case "CREATE_FAILED":
			return "", fmt.Errorf("stack creation failed")
		case "ROLLBACK_COMPLETE":
			return "", fmt.Errorf("stack creation failed.")
		case "DELETE_COMPLETE":
			return "", nil
		case "DELETE_FAILED":
			return "", fmt.Errorf("stack deletion failed.")
		}

		time.Sleep(2 * time.Second)
	}
}

func awsConfig(region string, creds *AwsCredentials) *aws.Config {
	config := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(creds.Access, creds.Secret, creds.Session),
	}

	if e := os.Getenv("AWS_ENDPOINT"); e != "" {
		config.Endpoint = aws.String(e)
	}

	return config
}
