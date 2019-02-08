package google

import (
	"fmt"
	"io"

	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
	kube "github.com/datacol-io/datacol/k8s"
)

func (g *GCPCloud) ProcessRun(name string, stream io.ReadWriter, opts pb.ProcessRunOptions) (string, error) {
	ns := g.DeploymentName
	cfg, err := getKubeClientConfig(ns)
	if err != nil {
		return "", err
	}

	app, _ := g.AppGet(name)
	envVars, _ := g.EnvironmentGet(name)
	sqlproxy := g.appLinkedDB(app)

	return kube.ProcessRun(g.kubeClient(), cfg, ns, name, g.latestImage(app), opts, envVars, sqlproxy, stream, cloud.GCPProvider)
}

func (g *GCPCloud) ProcessList(app string) ([]*pb.Process, error) {
	return kube.ProcessList(g.kubeClient(), g.DeploymentName, app)
}

func (g *GCPCloud) ProcessLimits(name, resource string, limits map[string]string) error {
	return kube.ProcessLimits(g.kubeClient(), g.DeploymentName, name, resource, limits)
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
