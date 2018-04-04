package local

import (
	"fmt"
	"io"
	"os"

	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/cloud/common"
	sched "github.com/datacol-io/datacol/cloud/kube"
)

func (g *LocalCloud) K8sConfigPath() (string, error) {
	return os.Getenv("HOME") + "/.kube/config", nil
}

func (g *LocalCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	return sched.LogStreamReq(g.kubeClient(), w, g.Name, app, opts)
}

func (g *LocalCloud) ProcessRun(name string, stream io.ReadWriter, command []string) error {
	ns := g.Name
	cfg, err := getKubeClientConfig(ns)
	if err != nil {
		return err
	}

	app, _ := g.AppGet(name)
	envVars, _ := g.EnvironmentGet(name)

	return sched.ProcessRun(g.kubeClient(), cfg, ns, name, g.latestImage(app), command, envVars, false, stream, cloud.LocalProvider)
}

func (g *LocalCloud) ProcessList(app string) ([]*pb.Process, error) {
	return sched.ProcessList(g.kubeClient(), g.Name, app)
}

func (g *LocalCloud) ProcessSave(name string, structure map[string]int32) error {
	app, err := g.AppGet(name)
	if err != nil {
		return err
	}

	build, err := g.BuildGet(app.Name, app.BuildId)
	if err != nil {
		return err
	}

	envVars, _ := g.EnvironmentGet(name)

	return common.ScaleApp(g.kubeClient(), g.Name, name, g.latestImage(app), envVars, false, build.Procfile, structure, cloud.LocalProvider)
}

func (g *LocalCloud) latestImage(app *pb.App) string {
	return fmt.Sprintf("%v/%v:%v", g.RegistryAddress, app.Name, app.BuildId)
}
