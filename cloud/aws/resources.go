package aws

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	pb "github.com/dinesh/datacol/api/models"
)

var notImplemented = fmt.Errorf("Not Implemented yet.")

func (a *AwsCloud) ResourceGet(name string) (*pb.Resource, error) {
	s, err := a.describeStack()
	if err != nil {
		return nil, err
	}
	rs := resourceFromStack(s)

	if rs.Tags["StackName"] != "" && rs.Tags["StackName"] != a.DeploymentName {
		return nil, fmt.Errorf("no such stack on this rack: %s", name)
	}

	switch rs.Tags["Resource"] {
	case "mysql":
		rs.Exports["URL"] = fmt.Sprintf("mysql://%s:%s@%s:%s/%s", rs.Outputs["EnvMysqlUsername"], rs.Outputs["EnvMysqlPassword"], rs.Outputs["Port3306TcpAddr"], rs.Outputs["Port3306TcpPort"], rs.Outputs["EnvMysqlDatabase"])
	case "postgres":
		rs.Exports["URL"] = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", rs.Outputs["EnvPostgresUsername"], rs.Outputs["EnvPostgresPassword"], rs.Outputs["Port5432TcpAddr"], rs.Outputs["Port5432TcpPort"], rs.Outputs["EnvPostgresDatabase"])
	default:
		return nil, fmt.Errorf("invaid Resource type %s", rs.Tags["Resource"])
	}

	return rs, nil
}

func (a *AwsCloud) ResourceDelete(name string) error {
	return notImplemented
}

func (a *AwsCloud) ResourceList() (pb.Resources, error) {
	return nil, notImplemented
}

func (a *AwsCloud) ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error) {
	rs := &pb.Resource{
		Name:       name,
		Kind:       kind,
		Parameters: cfParams(params),
	}

	var req *cloudformation.CreateStackInput
	var err error

	switch rs.Kind {
	case "mysql", "postgres":
		req, err = a.createResource(rs)
	default:
		return nil, fmt.Errorf("Unsupported resource type: %s", rs.Kind)
	}

	if err != nil {
		return nil, err
	}

	tags := map[string]string{
		"Name":      rs.Name,
		"Resource":  rs.Kind,
		"StackName": a.DeploymentName,
		"Type":      "resource",
	}

	fmt.Printf("tags: %+v reqtags: %+v\n", tags, req.Tags)

	for key, value := range tags {
		req.Tags = append(req.Tags, &cloudformation.Tag{Key: aws.String(key), Value: aws.String(value)})
	}

	fmt.Printf("reqtags: %+v\n", req.Tags)

	_, err = a.cloudformation().CreateStack(req)

	return rs, err
}

func (a *AwsCloud) ResourceLink(app, name string) (*pb.Resource, error) {
	return nil, notImplemented
}

func (a *AwsCloud) ResourceUnlink(app, name string) (*pb.Resource, error) {
	return nil, notImplemented
}

func (a *AwsCloud) createResource(s *pb.Resource) (*cloudformation.CreateStackInput, error) {
	formation, err := resourceFormation(s.Kind, nil)
	if err != nil {
		return nil, err
	}

	if err := a.appendSystemParameters(s); err != nil {
		return nil, err
	}

	if err := filterFormationParameters(s, formation); err != nil {
		return nil, err
	}

	req := &cloudformation.CreateStackInput{
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		StackName:    aws.String(fmt.Sprintf("%s-%s", a.DeploymentName, s.Name)),
		TemplateBody: aws.String(formation),
	}

	return req, nil
}

func resourceFromStack(s *cloudformation.Stack) *pb.Resource {
	tags := stackTags(s)
	name := coalesceString(tags["Name"], *s.StackName)
	exports := map[string]string{}
	params := stackParameters(s)

	if url, ok := params["Url"]; ok {
		exports["URL"] = url
	}

	rtype := tags["Resource"]
	if rtype == "" {
		rtype = tags["Service"]
	}

	return &pb.Resource{
		Name:    name,
		Kind:    rtype,
		Status:  *s.StackStatus,
		Tags:    tags,
		Outputs: stackOutputs(s),
		Exports: exports,
	}
}

func (p *AwsCloud) appendSystemParameters(s *pb.Resource) error {
	password := generatePassword()

	if s.Parameters["Password"] == "" {
		s.Parameters["Password"] = password
	}
	s.Parameters["SecurityGroups"] = os.Getenv("AWS_SECURITY_GROUP")
	s.Parameters["Subnets"] = os.Getenv("AWS_SUBNETS")
	s.Parameters["SubnetsPrivate"] = coalesceString(os.Getenv("AWS_SUBNET_PRIVATE"), os.Getenv("AWS_SUBNET"))
	s.Parameters["Vpc"] = os.Getenv("AWS_VPC_ID")
	// s.Parameters["VpcCidr"] = p.VpcCidr

	return nil
}

func stackTags(stack *cloudformation.Stack) map[string]string {
	tags := make(map[string]string)

	for _, tag := range stack.Tags {
		tags[*tag.Key] = *tag.Value
	}

	return tags
}

func coalesceString(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func filterFormationParameters(s *pb.Resource, formation string) error {
	var params struct {
		Parameters map[string]interface{}
	}

	if err := json.Unmarshal([]byte(formation), &params); err != nil {
		return err
	}

	for key := range s.Parameters {
		if _, ok := params.Parameters[key]; !ok {
			delete(s.Parameters, key)
		}
	}

	return nil
}

func resourceFormation(kind string, data interface{}) (string, error) {
	d, err := buildTemplate(kind, "resource", data)
	if err != nil {
		return "", err
	}

	return d, nil
}
