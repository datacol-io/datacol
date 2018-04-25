package local

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
	sched "github.com/datacol-io/datacol/k8s"
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
	var app *pb.App
	for _, a := range g.Apps {
		if a.Name == name {
			app = a
		}
	}

	if app == nil {
		return nil, fmt.Errorf("App Not Found")
	}

	if app.BuildId != "" {
		b, err := g.BuildGet(name, app.BuildId)
		if err != nil {
			return nil, err
		}

		proctype, kc := common.GetDefaultProctype(b), g.kubeClient()
		serviceName := common.GetJobID(name, proctype)

		if app.Endpoint, err = sched.GetServiceEndpoint(kc, g.Name, serviceName); err != nil {
			return app, err
		}
		return app, g.saveApp(app)
	}

	return app, nil
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
	sched.DeleteApp(g.kubeClient(), g.Name, name, cloud.LocalProvider)

	for i, a := range g.Apps {
		if a.Name == name {
			g.Apps = append(g.Apps[:i], g.Apps[i+1:]...)
		}
	}

	return nil
}

// DomainUpdate updates list of Domains for an app
// domain can be example.com if you want to add or :example.com if you want to delete
func (g *LocalCloud) AppUpdateDomain(name, domain string) error {
	app, err := g.AppGet(name)
	if err != nil {
		return err
	}

	app.Domains = common.MergeAppDomains(app.Domains, domain)

	return g.saveApp(app)
}
