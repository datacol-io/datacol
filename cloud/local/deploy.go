package local

import (
	"fmt"
	"io"
	"os"

	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud/common"
	sched "github.com/dinesh/datacol/cloud/kube"
)

func (g *LocalCloud) K8sConfigPath() (string, error) {
	return os.Getenv("HOME") + "/.kube/config", nil
}

func (g *LocalCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	c, err := getKubeClientset(g.Name)
	if err != nil {
		return err
	}

	return sched.LogStreamReq(c, w, g.Name, app, opts)
}

func (g *LocalCloud) ProcessRun(name string, stream io.ReadWriter, command string) error {
	ns := g.Name
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

	return sched.ProcessExec(c, cfg, ns, name, g.latestImage(app), command, envVars, stream)
}

func (g *LocalCloud) ProcessList(app string) ([]*pb.Process, error) {
	ns := g.Name
	c, err := getKubeClientset(ns)
	if err != nil {
		return nil, err
	}

	return sched.ProcessList(c, g.Name, app)
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

	c, err := getKubeClientset(g.Name)
	if err != nil {
		return err
	}

	return common.ScaleApp(c, g.Name, name, g.latestImage(app), build.Procfile, structure)
}

func (g *LocalCloud) latestImage(app *pb.App) string {
	return fmt.Sprintf("%v/%v:%v", g.RegistryAddress, app.Name, app.BuildId)
}
