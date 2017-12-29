package google

import (
	"fmt"
	"io"

	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud/kube"
)

func (g *GCPCloud) ProcessRun(name string, stream io.ReadWriter, command string) error {
	ns := g.DeploymentName
	cfg, err := getKubeClientConfig(ns)
	if err != nil {
		return err
	}

	c, err := getKubeClientset(ns)
	if err != nil {
		return err
	}
	app, _ := g.AppGet(name)
	envVars, _ := g.EnvironmentGet(name)

	return kube.ProcessExec(c, cfg, ns, name, g.latestImage(app), command, envVars, stream)
}

func (g *GCPCloud) latestImage(app *pb.App) string {
	return fmt.Sprintf("gcr.io/%v/%v:%v", g.Project, app.Name, app.BuildId)
}
