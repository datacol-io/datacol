package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
)

var notImplemented = fmt.Errorf("Not Implemented yet.")

func (a *AwsCloud) ResourceGet(name string) (*pb.Resource, error) {
	cfname := fmt.Sprintf("%s-%s", a.DeploymentName, name)
	s, err := a.describeStack(cfname)
	if err != nil {
		return nil, err
	}
	rs := resourceFromStack(s)

	if rs.Tags["Case"] != "" && rs.Tags["Case"] != a.DeploymentName {
		return nil, fmt.Errorf("no such stack on this case: %s", name)
	}

	if strings.HasSuffix(rs.Status, "FAILED") {
		a.setStatusReason(rs, cfname)
	}

	switch rs.Tags["Resource"] {
	case "mysql":
		rs.Exports["URL"] = fmt.Sprintf("mysql://%s:%s@%s:%s/%s", rs.Outputs["EnvMysqlUsername"], rs.Outputs["EnvMysqlPassword"], rs.Outputs["Port3306TcpAddr"], rs.Outputs["Port3306TcpPort"], rs.Outputs["EnvMysqlDatabase"])
	case "postgres":
		rs.Exports["URL"] = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", rs.Outputs["EnvPostgresUsername"], rs.Outputs["EnvPostgresPassword"], rs.Outputs["Port5432TcpAddr"], rs.Outputs["Port5432TcpPort"], rs.Outputs["EnvPostgresDatabase"])
	case "redis":
		rs.Exports["URL"] = fmt.Sprintf("redis://%s:%s/%s", rs.Outputs["Port6379TcpAddr"], rs.Outputs["Port6379TcpPort"], rs.Outputs["EnvRedisDatabase"])
	case "elasticsearch":
		rs.Exports["DOMAIN_URL"] = rs.Outputs["EnvEndpoint"]
		rs.Exports["DOMAIN_ARN"] = rs.Outputs["EnvArn"]
	}

	return rs, nil
}

func (a *AwsCloud) ResourceDelete(name string) error {
	s, err := a.ResourceGet(name)
	if err != nil {
		return err
	}

	apps, err := a.resourceApps(s)
	if err != nil {
		return err
	}

	if len(apps) > 0 {
		return fmt.Errorf("resource is linked to %s", apps[0].Name)
	}

	_, err = a.cloudformation().DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(s.Stack),
	})

	return err
}

func (a *AwsCloud) ResourceList() (pb.Resources, error) {
	stacks, err := a.describeStacks(&cloudformation.DescribeStacksInput{})
	if err != nil {
		return nil, err
	}

	resources := pb.Resources{}

	for _, stack := range stacks {
		tags := stackTags(stack)

		if tags["System"] == "datacol" && (tags["Type"] == "resource" || tags["Type"] == "app") {
			if tags["Case"] == a.DeploymentName || tags["Case"] == "" {
				resources = append(resources, resourceFromStack(stack))
			}
		}
	}

	for _, s := range resources {
		apps, err := a.resourceApps(s)
		if err != nil {
			return nil, err
		}
		for _, a := range apps {
			s.Apps = append(s.Apps, a.Name)
		}
	}

	return resources, nil
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
	case "mysql", "postgres", "redis", "app", "elasticsearch":
		req, err = a.createResource(rs)
	default:
		err = fmt.Errorf("Unsupported resource type: %s", rs.Kind)
	}

	if err != nil {
		return nil, err
	}

	for key := range rs.Parameters {
		req.Parameters = append(req.Parameters, &cloudformation.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(rs.Parameters[key]),
		})
	}

	tags := map[string]string{
		"Name":     rs.Name,
		"Resource": rs.Kind,
		"Case":     a.DeploymentName,
		"Type":     "resource",
		"System":   "datacol",
	}

	for key, value := range tags {
		req.Tags = append(req.Tags, &cloudformation.Tag{Key: aws.String(key), Value: aws.String(value)})
	}

	_, err = a.cloudformation().CreateStack(req)

	return rs, err
}

func (p *AwsCloud) ResourceLink(app, name string) (*pb.Resource, error) {
	a, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	s, err := p.ResourceGet(name)
	if err != nil {
		return nil, err
	}

	// already linked
	for _, linkedApp := range s.Apps {
		if a.Name == linkedApp {
			return nil, fmt.Errorf("resource %s is already linked to app %s", s.Name, a.Name)
		}
	}

	return s, err
}

func (p *AwsCloud) ResourceUnlink(app, name string) (*pb.Resource, error) {
	a, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	s, err := p.ResourceGet(name)
	if err != nil {
		return nil, err
	}

	// already linked
	linked := false
	for _, linkedApp := range s.Apps {
		if a.Name == linkedApp {
			linked = true
			break
		}
	}

	if !linked {
		return nil, fmt.Errorf("resource %s is not linked to app %s", s.Name, a.Name)
	}

	return s, err
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

	log.Debugf("creating resource with %s", toJson(s.Parameters))

	req := &cloudformation.CreateStackInput{
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		StackName:    aws.String(fmt.Sprintf("%s-%s", a.DeploymentName, s.Name)),
		TemplateBody: aws.String(formation),
	}

	return req, nil
}

func (p *AwsCloud) resourceApps(s *pb.Resource) (pb.Apps, error) {
	return pb.Apps{}, nil
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
		Outputs: stackOutputs(s),
		Exports: exports,
		Stack:   *s.StackName,
		Status:  *s.StackStatus,
		Tags:    tags,
	}
}

func (p *AwsCloud) appendSystemParameters(s *pb.Resource) error {
	password := generatePassword()

	if s.Parameters["Password"] == "" {
		s.Parameters["Password"] = password
	}
	s.Parameters["SecurityGroups"] = os.Getenv("AWS_SECURITY_GROUP")
	s.Parameters["Subnets"] = os.Getenv("AWS_SUBNETS")
	s.Parameters["SubnetsPrivate"] = coalesceString(os.Getenv("AWS_SUBNETS_PRIVATE"), os.Getenv("AWS_SUBNETS"))

	// Expose the public and private subnet since some of the resource only accept single subnet
	if s.Parameters["Subnets"] != "" {
		s.Parameters["Subnet"] = strings.Split(s.Parameters["Subnets"], ",")[0]
	}

	if s.Parameters["SubnetsPrivate"] != "" {
		s.Parameters["SubnetPrivate"] = strings.Split(s.Parameters["SubnetsPrivate"], ",")[0]
	}

	s.Parameters["Vpc"] = os.Getenv("AWS_VPC_ID")
	s.Parameters["VpcCidr"] = os.Getenv("AWS_VPC_CIDR")

	return nil
}

func (p *AwsCloud) setStatusReason(r *pb.Resource, stackName string) error {
	eres, err := p.describeStackEvents(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return err
	}

	for _, event := range eres.StackEvents {
		if *event.ResourceStatus == cloudformation.ResourceStatusCreateFailed ||
			*event.ResourceStatus == cloudformation.ResourceStatusDeleteFailed {
			r.StatusReason = *event.ResourceStatusReason
			break
		}
	}
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

// filterFormationParameters will filter the parameters not defined in CF template
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
