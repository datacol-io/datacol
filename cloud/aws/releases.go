package aws

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
	"k8s.io/client-go/pkg/util/intstr"
	"os"
	"strconv"
)

func (a *AwsCloud) dynamoReleases() string {
	return fmt.Sprintf("%s-releases", a.DeploymentName)
}

func (a *AwsCloud) ReleaseList(app string, limit int) (pb.Releases, error) {
	return nil, nil
}

func (a *AwsCloud) ReleaseDelete(app, id string) error {
	return nil
}

func (a *AwsCloud) BuildRelease(b *pb.Build) (*pb.Release, error) {
	image := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s-%s", os.Getenv("AWS_ACCOUNT_ID"), a.Region, a.ecrRepository(), b.App, b.Id)
	log.Debugf("---- Docker Image: %s", image)

	envVars, err := a.EnvironmentGet(b.App)
	if err != nil {
		return nil, err
	}
	c, err := getKubeClientset(a.DeploymentName)
	if err != nil {
		return nil, err
	}

	deployer, err := sched.NewDeployer(c)
	if err != nil {
		return nil, err
	}

	port := 8080
	if pv, ok := envVars["PORT"]; ok {
		p, err := strconv.Atoi(pv)
		if err != nil {
			return nil, err
		}
		port = p
	}

	ret, err := deployer.Run(&sched.DeployRequest{
		ServiceID:     b.App,
		Image:         image,
		Replicas:      1,
		Environment:   a.DeploymentName,
		Zone:          a.Region,
		ContainerPort: intstr.FromInt(port),
		EnvVars:       envVars,
	})

	if err != nil {
		return nil, err
	}

	log.Debugf("Deploying %s with %s", b.App, toJson(ret.Request))

	r := &pb.Release{
		Id:        generateId("R", 5),
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
	}

	return r, a.releaseSave(r)
}

func (a *AwsCloud) releaseSave(r *pb.Release) error {
	req := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id":       {S: aws.String(r.Id)},
			"app":      {S: aws.String(r.App)},
			"build_id": {S: aws.String(r.BuildId)},
		},
		TableName: aws.String(a.dynamoReleases()),
	}

	if r.Status != "" {
		req.Item["status"] = &dynamodb.AttributeValue{S: aws.String(r.Status)}
	}

	_, err := a.dynamodb().PutItem(req)
	return err
}
