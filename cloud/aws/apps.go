package aws

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ecr"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
	sched "github.com/datacol-io/datacol/k8s"
)

func (a *AwsCloud) AppList() (pb.Apps, error) {
	req := &dynamodb.ScanInput{
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(a.dynamoApps()),
	}

	res, err := a.dynamodb().Scan(req)
	if err != nil {
		return nil, err
	}

	apps := make(pb.Apps, len(res.Items))

	for i, item := range res.Items {
		apps[i] = a.appFromItem(item)
	}

	return apps, nil
}

func (a *AwsCloud) AppCreate(name string, req *pb.AppCreateOptions) (*pb.App, error) {
	if _, err := a.AppGet(name); err == nil {
		return nil, fmt.Errorf("Duplicate app: %s", name)
	}

	if _, err := a.ResourceCreate(stackNameForApp(name), "app", map[string]string{
		"AppName":       name,
		"StackName":     a.DeploymentName,
		"BucketPrefix":  fmt.Sprintf("%s/%s", a.SettingBucket, name),
		"RepositoryUrl": req.RepoUrl,
	}); err != nil {
		return nil, fmt.Errorf("creating environment for %s. err: %v", name, err)
	}

	app := &pb.App{Name: name, Status: pb.StatusCreated}
	return app, a.saveApp(app)
}

func (a *AwsCloud) AppRestart(app string) error {
	log.Debugf("Restarting %s", app)
	env, err := a.EnvironmentGet(app)
	if err != nil {
		return err
	}

	env["_RESTARTED"] = time.Now().Format("20060102150405")
	return sched.SetPodEnv(a.kubeClient(), a.DeploymentName, app, env)
}

func (a *AwsCloud) AppGet(name string) (*pb.App, error) {
	cfname := cfNameForApp(a.DeploymentName, name)
	if _, err := a.describeStack(cfname); err != nil {
		return nil, fmt.Errorf("no such app found: %s", name)
	}

	req := &dynamodb.GetItemInput{
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"name": {S: aws.String(name)},
		},
		TableName: aws.String(a.dynamoApps()),
	}

	res, err := a.dynamodb().GetItem(req)
	if err != nil {
		return nil, fmt.Errorf("fetching from dynamodb err: %v", err)
	}

	if res.Item == nil {
		return nil, fmt.Errorf("no such app found by name: %s", name)
	}

	app := a.appFromItem(res.Item)

	if app.BuildId != "" {
		b, err := a.BuildGet(name, app.BuildId)
		if err != nil {
			return nil, err
		}

		proctype, kc := common.GetDefaultProctype(b), a.kubeClient()
		serviceName := common.GetJobID(name, proctype)

		if app.Endpoint, err = sched.GetServiceEndpoint(kc, a.DeploymentName, serviceName); err != nil {
			return app, err
		}
		return app, a.saveApp(app)
	}

	return app, nil
}

func (a *AwsCloud) AppDelete(name string) error {
	a.deleteFromCluster(name)

	if err := a.deleteFromDynamo(name); err != nil {
		return err
	}

	if err := a.deleteAppResources(name); err != nil {
		return err
	}

	if err := a.ResourceDelete(stackNameForApp(name)); err != nil {
		return err
	}

	return nil
}

func (p *AwsCloud) deleteAppResources(name string) error {
	svc := p.ecr()

	out, err := svc.ListImages(&ecr.ListImagesInput{
		RegistryId:     aws.String(os.Getenv("AWS_ACCOUNT_ID")),
		RepositoryName: aws.String(p.ecrRepository(name)),
	})
	if err != nil {
		return err
	}

	log.Debugf("Deleting docker images: %d ...", len(out.ImageIds))

	if len(out.ImageIds) > 0 {
		_, err = svc.BatchDeleteImage(&ecr.BatchDeleteImageInput{
			RegistryId:     aws.String(os.Getenv("AWS_ACCOUNT_ID")),
			RepositoryName: aws.String(p.ecrRepository(name)),
			ImageIds:       out.ImageIds,
		})
		return err
	}

	return nil
}

func (p *AwsCloud) deleteFromCluster(name string) error {
	log.Debugf("Removing app from kube cluster ...")
	return sched.DeleteApp(p.kubeClient(), p.DeploymentName, name, cloud.AwsProvider)
}

func (p *AwsCloud) deleteFromDynamo(name string) error {
	_, err := p.dynamodb().DeleteItem(&dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {S: aws.String(name)},
		},
		TableName: aws.String(p.dynamoApps()),
	})
	return err
}

func (p *AwsCloud) saveApp(a *pb.App) error {
	req := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"name": {S: aws.String(a.Name)},
		},
		TableName: aws.String(p.dynamoApps()),
	}

	if a.Status != "" {
		req.Item["status"] = &dynamodb.AttributeValue{S: aws.String(a.Status)}
	}

	if a.Endpoint != "" {
		req.Item["endpoint"] = &dynamodb.AttributeValue{S: aws.String(a.Endpoint)}
	}

	if a.BuildId != "" {
		req.Item["build_id"] = &dynamodb.AttributeValue{S: aws.String(a.BuildId)}
	}

	if a.ReleaseId != "" {
		req.Item["release_id"] = &dynamodb.AttributeValue{S: aws.String(a.ReleaseId)}
	}

	if len(a.Domains) > 0 {
		list := []*dynamodb.AttributeValue{}
		for _, d := range a.Domains {
			list = append(list, &dynamodb.AttributeValue{S: aws.String(d)})
		}

		req.Item["domains"] = &dynamodb.AttributeValue{L: list}
	}

	_, err := p.dynamodb().PutItem(req)
	return err
}

func (a *AwsCloud) dynamoApps() string {
	return fmt.Sprintf("%s-apps", a.DeploymentName)
}

func (a *AwsCloud) appFromItem(item map[string]*dynamodb.AttributeValue) *pb.App {
	name := coalesce(item["name"], "")

	app := &pb.App{
		Name:      name,
		Status:    coalesce(item["status"], ""),
		ReleaseId: coalesce(item["release_id"], ""),
		BuildId:   coalesce(item["build_id"], ""),
		Endpoint:  coalesce(item["endpoint"], ""),
	}

	if domainValues, ok := item["domains"]; ok {
		domains := []string{}
		for _, key := range domainValues.L {
			domains = append(domains, coalesce(key, ""))
		}
		app.Domains = domains
	}

	return app
}

func stackNameForApp(a string) string {
	return fmt.Sprintf("app-%s", a)
}

func cfNameForApp(a, b string) string {
	return fmt.Sprintf("%s-app-%s", a, b)
}
