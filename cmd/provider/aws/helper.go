package aws

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	log "github.com/Sirupsen/logrus"
	term "github.com/appscode/go-term"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"io/ioutil"
	"os"
	"strings"
)

var events = map[string]bool{}

func displayProgress(stack string, cf *cloudformation.CloudFormation, isDeleting bool) error {
	res, err := cf.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stack),
	})

	if err != nil {
		return err
	}

	for _, event := range res.StackEvents {
		if events[*event.EventId] {
			continue
		}

		events[*event.EventId] = true

		// Log all CREATE_FAILED to display
		if !isDeleting && *event.ResourceStatus == "CREATE_FAILED" {
			msg := fmt.Sprintf("Failed %s: %s", *event.ResourceType, *event.ResourceStatusReason)
			fmt.Println(msg)
		}

		name := friendlyName(*event.ResourceType)

		if name == "" {
			continue
		}

		switch *event.ResourceStatus {
		case "CREATE_IN_PROGRESS":
		case "CREATE_COMPLETE":
			if !isDeleting {
				id := *event.PhysicalResourceId

				if strings.HasPrefix(id, "arn:") {
					id = *event.LogicalResourceId
				}

				fmt.Printf("Created %s: %s\n", name, id)
			}
		case "CREATE_FAILED":
		case "DELETE_IN_PROGRESS":
		case "DELETE_COMPLETE":
			id := *event.PhysicalResourceId

			if strings.HasPrefix(id, "arn:") {
				id = *event.LogicalResourceId
			}

			fmt.Printf("Deleted %s: %s\n", name, id)
		case "DELETE_SKIPPED":
			id := *event.PhysicalResourceId

			if strings.HasPrefix(id, "arn:") {
				id = *event.LogicalResourceId
			}

			fmt.Printf("Skipped %s: %s\n", name, id)
		case "DELETE_FAILED":
			id := *event.PhysicalResourceId

			if strings.HasPrefix(id, "arn:") {
				id = *event.LogicalResourceId
			}

			fmt.Printf("Failed to delete %s: %s\n", name, id)
		case "ROLLBACK_IN_PROGRESS", "ROLLBACK_COMPLETE":
		case "UPDATE_IN_PROGRESS", "UPDATE_COMPLETE", "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS", "UPDATE_FAILED", "UPDATE_ROLLBACK_IN_PROGRESS", "UPDATE_ROLLBACK_COMPLETE", "UPDATE_ROLLBACK_FAILED":
		default:
			return fmt.Errorf("Unhandled status: %s\n", *event.ResourceStatus)
		}
	}

	return nil
}

func ReadCredentials(fileName string) (creds *AwsCredentials, err error) {
	// read credentials from ENV
	creds = &AwsCredentials{
		Access:  os.Getenv("AWS_ACCESS_KEY_ID"),
		Secret:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Session: os.Getenv("AWS_SESSION_TOKEN"),
	}

	// if filename argument provided, prefer these credentials over any found in the environment
	var inputCreds *AwsCredentials
	if fileName != "" {
		inputCreds, err = readCredentialsFromFile(fileName)
	}

	if err != nil {
		return nil, err
	}

	if inputCreds != nil {
		creds = inputCreds
	}

	if creds.Access == "" || creds.Secret == "" {
		if creds.Access == "" {
			creds.Access = term.Read("AWS Access Key ID: ")
		}

		if creds.Secret == "" {
			creds.Secret = term.Read("AWS Secret Access Key: ")
		}

		fmt.Println("")
	}

	creds.Access = strings.TrimSpace(creds.Access)
	creds.Secret = strings.TrimSpace(creds.Secret)
	creds.Session = strings.TrimSpace(creds.Session)

	return
}

func readCredentialsFromFile(credentialsCsvFileName string) (*AwsCredentials, error) {
	fmt.Printf("Reading credentials from file %s\n", credentialsCsvFileName)
	credsFile, err := ioutil.ReadFile(credentialsCsvFileName)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(credsFile))

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	creds := &AwsCredentials{}
	if len(records) == 2 {
		switch len(records[0]) {

		case 2:
			// Access key ID,Secret access key
			creds.Access = records[1][0]
			creds.Secret = records[1][1]

		case 3:
			// User name,Access key ID,Secret access key
			creds.Access = records[1][1]
			creds.Secret = records[1][2]

		case 5:
			// User name,Password,Access key ID,Secret access key,Console login link
			creds.Access = records[1][2]
			creds.Secret = records[1][3]

		default:
			return creds, fmt.Errorf("credentials secrets is of unknown length")
		}
	} else {
		return creds, fmt.Errorf("credentials file is of unknown length")
	}

	return creds, nil
}

// FriendlyName turns an AWS resource type into a friendly name
func friendlyName(t string) string {
	switch t {
	case "AWS::AutoScaling::AutoScalingGroup":
		return "AutoScalingGroup"
	case "AWS::AutoScaling::LaunchConfiguration":
		return ""
	case "AWS::AutoScaling::LifecycleHook":
		return ""
	case "AWS::CloudFormation::Stack":
		return "CloudFormation Stack"
	case "AWS::DynamoDB::Table":
		return "DynamoDB Table"
	case "AWS::EC2::EIP":
		return "NAT Elastic IP"
	case "AWS::EC2::InternetGateway":
		return "VPC Internet Gateway"
	case "AWS::EC2::Route":
		return ""
	case "AWS::EC2::RouteTable":
		return "Routing Table"
	case "AWS::EC2::SecurityGroup":
		return "Security Group"
	case "AWS::EC2::Subnet":
		return "VPC Subnet"
	case "AWS::EC2::SubnetRouteTableAssociation":
		return ""
	case "AWS::EC2::VPC":
		return "VPC"
	case "AWS::EC2::VPCGatewayAttachment":
		return ""
	case "AWS::EC2::Instance":
		return "EC2 Instance"
	case "AWS::ECS::Cluster":
		return "ECS Cluster"
	case "AWS::ECS::Service":
		return "ECS Service"
	case "AWS::EFS::FileSystem":
		return "EFS Filesystem"
	case "AWS::EFS::MountTarget":
		return ""
	case "AWS::ElasticLoadBalancing::LoadBalancer":
		return "Elastic Load Balancer"
	case "AWS::Events::Rule":
		return ""
	case "AWS::IAM::AccessKey":
		return "Access Key"
	case "AWS::IAM::InstanceProfile":
		return ""
	case "AWS::IAM::ManagedPolicy":
		return "IAM Managed Policy"
	case "AWS::IAM::Role":
		return ""
	case "AWS::IAM::User":
		return "IAM User"
	case "AWS::Kinesis::Stream":
		return "Kinesis Stream"
	case "AWS::KMS::Alias":
		return "KMS Alias"
	case "AWS::Lambda::Function":
		return "Lambda Function"
	case "AWS::Lambda::Permission":
		return ""
	case "AWS::Logs::LogGroup":
		return "CloudWatch Log Group"
	case "AWS::Logs::SubscriptionFilter":
		return ""
	case "AWS::EC2::NatGateway":
		return "NAT Gateway"
	case "AWS::S3::Bucket":
		return "S3 Bucket"
	case "AWS::SNS::Topic":
		return ""
	case "AWS::SQS::Queue":
		return "SQS Queue"
	case "AWS::SQS::QueuePolicy":
		return ""
	case "Custom::EC2AvailabilityZones":
		return ""
	case "Custom::ECSService":
		return "ECS Service"
	case "Custom::ECSTaskDefinition":
		return "ECS TaskDefinition"
	case "Custom::KMSKey":
		return "KMS Key"
	case "AWS::EC2::VPCDHCPOptionsAssociation", "AWS::EC2::DHCPOptions":
		return ""
	}

	return fmt.Sprintf("Unknown: %s", t)
}

func prompt(s string) {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("%s\n\nPlease press [ENTER] or Ctrl-C to cancel", s)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if line == "\n" {
			break
		}
	}
}
