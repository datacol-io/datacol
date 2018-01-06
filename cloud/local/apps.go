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

	env, err := g.EnvironmentGet(app)
	if err != nil {
		return err
	}

	env["_RESTARTED"] = time.Now().Format("20060102150405")

	if err = sched.SetPodEnv(g.kubeClient(), ns, app, env); err != nil {
		return err
	}

	return nil
}

func (g *LocalCloud) AppGet(name string) (*pb.App, error) {
	for _, a := range g.Apps {
		if a.Name == name {
			return a, nil
		}
	}

	return nil, fmt.Errorf("App Not Found")
}

func (g *LocalCloud) saveApp(a *pb.App) error {
	for i, app := range g.Apps {
		if app.Name == a.Name {
			g.Apps[i] = a
			break
		}
	}

	return nil
}

func (g *LocalCloud) AppDelete(name string) error {
	ns := g.Name

	podNames, err := sched.GetAllPodNames(g.kubeClient(), ns, name)
	if err != nil {
		return err
	}

	for _, pod := range podNames {
		sched.DeleteService(g.kubeClient(), ns, pod)
	}

	for i, a := range g.Apps {
		if a.Name == name {
			g.Apps = append(g.Apps[:i], g.Apps[i+1:]...)
		}
	}

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
