package google

import (
	"context"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
	sched "github.com/datacol-io/datacol/k8s"
)

const appKind = "App"

func (g *GCPCloud) AppList() (pb.Apps, error) {
	var apps pb.Apps

	q := datastore.NewQuery(appKind).Namespace(g.DeploymentName)
	if _, err := g.datastore().GetAll(context.Background(), q, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}

func (g *GCPCloud) AppCreate(name string, req *pb.AppCreateOptions) (*pb.App, error) {
	app := &pb.App{Name: name, Status: pb.StatusCreated}
	ctx, key := g.nestedKey(appKind, name)
	_, err := g.datastore().Put(ctx, key, app)

	return app, err
}

func (g *GCPCloud) AppRestart(app string) error {
	log.Debugf("Restarting pods inside %s", app)
	ns := g.DeploymentName

	env, err := g.EnvironmentGet(app)
	if err != nil {
		return err
	}

	env["_RESTARTED"] = time.Now().Format("20060102150405")
	return sched.SetPodEnv(g.kubeClient(), ns, app, env)
}

func (g *GCPCloud) AppGet(name string) (*pb.App, error) {
	app := new(pb.App)

	ctx, key := g.nestedKey(appKind, name)
	if err := g.datastore().Get(ctx, key, app); err != nil {
		return nil, err
	}

	if app.BuildId != "" {
		b, err := g.BuildGet(name, app.BuildId)
		if err != nil {
			return nil, err
		}

		proctype := common.GetDefaultProctype(b)
		serviceName := common.GetJobID(name, proctype)
		endpoint, err := sched.GetServiceEndpoint(g.kubeClient(), g.DeploymentName, serviceName)
		if err != nil {
			return app, nil
		}
		app.Endpoint = endpoint

		_, err = g.datastore().Put(ctx, key, app)
		return app, err
	}

	return app, nil
}

func (g *GCPCloud) AppDelete(name string) error {
	g.deleteAppFromCluster(name)
	return g.deleteAppFromDatastore(name)
}

// DomainUpdate updates list of Domains for an app
// domain can be example.com if you want to add or :example.com if you want to delete
func (g *GCPCloud) AppUpdateDomain(name, domain string) error {
	app, err := g.AppGet(name)
	if err != nil {
		return err
	}

	app.Domains = common.MergeAppDomains(app.Domains, domain)

	return g.saveApp(app)
}

func (g *GCPCloud) saveApp(app *pb.App) error {
	ctx, key := g.nestedKey(appKind, app.Name)
	_, err := g.datastore().Put(ctx, key, app)
	return err
}

func (g *GCPCloud) appLinkedDB(app *pb.App) bool {
	return g.appLinkedWith(app, "mysql") || g.appLinkedWith(app, "postgres")
}

func (g *GCPCloud) appLinkedWith(app *pb.App, kind string) bool {
	for _, r := range app.Resources {
		parts := strings.Split(r, "-")
		if parts[0] == kind {
			return true
		}
	}

	return false
}

func (g *GCPCloud) deleteAppFromDatastore(name string) error {
	store, ctx := g.datastore(), context.Background()

	q := datastore.NewQuery(buildKind).Namespace(g.DeploymentName).Filter("app =", name).KeysOnly()
	if err := deleteFromQuery(store, ctx, q); err != nil {
		return err
	}

	q = datastore.NewQuery(releaseKind).Namespace(g.DeploymentName).Filter("app =", name).KeysOnly()

	if err := deleteFromQuery(store, ctx, q); err != nil {
		return err
	}

	ctx, key := g.nestedKey(appKind, name)
	if err := store.Delete(ctx, key); err != nil {
		return err
	}

	return nil
}

func (g *GCPCloud) deleteAppFromCluster(name string) error {
	return sched.DeleteApp(g.kubeClient(), g.DeploymentName, name, cloud.GCPProvider)
}
