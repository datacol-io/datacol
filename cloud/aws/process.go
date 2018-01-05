package aws

import (
	"fmt"
	"io"
	"os"

	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud/common"
	"github.com/dinesh/datacol/cloud/kube"
)

func (p *AwsCloud) ProcessList(app string) ([]*pb.Process, error) {
	return kube.ProcessList(p.kubeClient(), p.DeploymentName, app)
}

func (p *AwsCloud) ProcessRun(name string, r io.ReadWriter, command string) error {
	ns := p.DeploymentName
	cfg, _ := getKubeClientConfig(ns)

	envVars, _ := p.EnvironmentGet(name)
	app, _ := p.AppGet(name)

	return kube.ProcessExec(p.kubeClient(), cfg, ns, name, p.latestImage(app), command, envVars, r)
}

func (p *AwsCloud) ProcessSave(name string, structure map[string]int32) error {
	app, err := p.AppGet(name)
	if err != nil {
		return err
	}

	build, err := p.BuildGet(app.Name, app.BuildId)
	if err != nil {
		return err
	}

	return common.ScaleApp(p.kubeClient(), p.DeploymentName, name,
		p.latestImage(app), build.Procfile, structure)
}

func (p *AwsCloud) latestImage(app *pb.App) string {
	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		os.Getenv("AWS_ACCOUNT_ID"),
		p.Region,
		p.ecrRepository(app.Name),
		app.BuildId,
	)
}
