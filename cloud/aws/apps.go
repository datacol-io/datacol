package aws

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
	sched "github.com/datacol-io/datacol/k8s"
)

func (a *AwsCloud) AppList() (pb.Apps, error) {
	return a.store.AppList()
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
	return app, a.store.AppCreate(app, req)
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

	app, err := a.store.AppGet(name)
	if err != nil {
		return nil, fmt.Errorf("no such app found by name: %s", name)
	}

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
		return app, a.store.AppUpdate(app)
	}

	return app, nil
}

func (a *AwsCloud) AppDelete(name string) error {
	a.deleteFromCluster(name)

	if err := a.store.AppDelete(name); err != nil {
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

// DomainUpdate updates list of Domains for an app
// domain can be example.com if you want to add or :example.com if you want to delete
func (p *AwsCloud) AppUpdateDomain(name, domain string) error {
	app, err := p.AppGet(name)
	if err != nil {
		return err
	}

	app.Domains = common.MergeAppDomains(app.Domains, domain)

	return p.store.AppUpdate(app)
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

func stackNameForApp(a string) string {
	return fmt.Sprintf("app-%s", a)
}

func cfNameForApp(a, b string) string {
	return fmt.Sprintf("%s-app-%s", a, b)
}
