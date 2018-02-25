package google

import (
	"fmt"
	"io"

	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud"
	"github.com/dinesh/datacol/cloud/common"
	"github.com/dinesh/datacol/cloud/kube"
)

func (g *GCPCloud) ProcessRun(name string, stream io.ReadWriter, command string) error {
	ns := g.DeploymentName
	cfg, err := getKubeClientConfig(ns)
	if err != nil {
		return err
	}

	app, _ := g.AppGet(name)
	envVars, _ := g.EnvironmentGet(name)
	sqlproxy := g.appLinkedDB(app)

	return kube.ProcessExec(g.kubeClient(), cfg, ns, name, g.latestImage(app), command, envVars, sqlproxy, stream, cloud.GCPProvider)
}

func (g *GCPCloud) ProcessList(app string) ([]*pb.Process, error) {
	return kube.ProcessList(g.kubeClient(), g.DeploymentName, app)
}

func (g *GCPCloud) ProcessSave(name string, structure map[string]int32) error {
	app, err := g.AppGet(name)
	if err != nil {
		return err
	}

	build, err := g.BuildGet(app.Name, app.BuildId)
	if err != nil {
		return err
	}

	envVars, _ := g.EnvironmentGet(name)

	return common.ScaleApp(g.kubeClient(), g.DeploymentName, name, g.latestImage(app), envVars, g.appLinkedDB(app), build.Procfile, structure, cloud.GCPProvider)
}

func (g *GCPCloud) latestImage(app *pb.App) string {
	return fmt.Sprintf("gcr.io/%v/%v:%v", g.Project, app.Name, app.BuildId)
}
