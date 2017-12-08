package aws

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func (a *AwsCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	image := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		os.Getenv("AWS_ACCOUNT_ID"), a.Region, a.ecrRepository(b.App), b.Id,
	)
	log.Debugf("---- Docker Image: %s", image)

	app, err := a.AppGet(b.App)
	if err != nil {
		return nil, err
	}

	envVars, err := a.EnvironmentGet(b.App)
	if err != nil {
		return nil, err
	}

	c := a.kubeClient()

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

	domains := sched.MergeAppDomains(app.Domains, options.Domain)

	ret, err := deployer.Run(&sched.DeployRequest{
		ServiceID:     b.App,
		Image:         image,
		Replicas:      1,
		Environment:   a.DeploymentName,
		Zone:          a.Region,
		ContainerPort: intstr.FromInt(port),
		EnvVars:       envVars,
		Domains:       domains,
	})

	if err != nil {
		return nil, err
	}

	if len(app.Domains) != len(domains) {
		app.Domains = domains
		a.saveApp(app)
	}

	log.Debugf("Deployed %s with %s", b.App, toJson(ret.Request))

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
