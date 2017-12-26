package local

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/appscode/go/log"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud/common"
	sched "github.com/dinesh/datacol/cloud/kube"
)

func (l *LocalCloud) AppList() (pb.Apps, error) {
	return l.Apps, nil
}

func (g *LocalCloud) AppCreate(name string, req *pb.AppCreateOptions) (*pb.App, error) {
	if _, err := g.AppGet(name); err != nil {
		g.Apps = append(g.Apps, &pb.App{
			Name: name,
		})
	}

	return g.AppGet(name)
}

func (g *LocalCloud) AppRestart(app string) error {
	log.Debugf("Restarting %s", app)
	ns := g.Name

	kube, err := getKubeClientset(ns)
	if err != nil {
		return err
	}

	env, err := g.EnvironmentGet(app)
	if err != nil {
		return err
	}

	env["_RESTARTED"] = time.Now().Format("20060102150405")
	return sched.SetPodEnv(kube, ns, app, env)
}

func (g *LocalCloud) AppGet(name string) (*pb.App, error) {
	for _, a := range g.Apps {
		if a.Name == name {
			return a, nil
		}
	}

	return nil, fmt.Errorf("App Not Found")
}

func (g *LocalCloud) AppDelete(name string) error {
	return nil
}

func (g *LocalCloud) EnvironmentGet(name string) (pb.Environment, error) {
	return g.EnvMap[name], nil
}

func (g *LocalCloud) EnvironmentSet(name string, body io.Reader) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	g.EnvMap[name] = common.LoadEnvironment(data)
	return nil
}

func (g *LocalCloud) GetRunningPods(app string) (string, error) {
	ns := g.Name
	c, err := getKubeClientset(ns)
	if err != nil {
		return "", err
	}

	return sched.RunningPods(ns, app, c)
}
