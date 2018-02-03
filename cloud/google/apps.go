package google

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud/common"
	sched "github.com/dinesh/datacol/cloud/kube"
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
	log.Debugf("Restarting %s", app)
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

	b, err := g.BuildGet(name, app.BuildId)
	if err != nil {
		return nil, err
	}

	// FIXME: Have a better way to determine name for the deployed service(s). This will change if we support multiple processes for an app.
	var proctype string
	if len(b.Procfile) > 0 {
		proctype = "web"
	} else {
		proctype = "cmd"
	}

	serviceName := common.GetJobID(name, proctype)
	endpoint, err := sched.GetServiceEndpoint(g.kubeClient(), g.DeploymentName, serviceName)
	if err != nil {
		return app, nil
	}
	app.Endpoint = endpoint

	_, err = g.datastore().Put(ctx, key, app)
	return app, err
}

func (g *GCPCloud) AppDelete(name string) error {
	g.deleteAppFromCluster(name)
	return g.deleteAppFromDatastore(name)
}

func (g *GCPCloud) saveApp(app *pb.App) error {
	ctx, key := g.nestedKey(appKind, app.Name)
	_, err := g.datastore().Put(ctx, key, app)
	return err
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
	return sched.DeleteService(g.kubeClient(), g.DeploymentName, name)
}
