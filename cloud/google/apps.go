package google

import (
	"time"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
	kerrors "k8s.io/client-go/pkg/api/errors"
	kapi "k8s.io/client-go/pkg/api/v1"
	klabels "k8s.io/client-go/pkg/labels"
)

const appKind = "App"

func (g *GCPCloud) AppList() (pb.Apps, error) {
	var apps pb.Apps

	q := datastore.NewQuery(appKind)
	if _, err := g.datastore().GetAll(g.ctxNS(), q, &apps); err != nil {
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

func (g *GCPCloud) AppGet(name string) (*pb.App, error) {
	app := new(pb.App)

	ctx, key := g.nestedKey(appKind, name)
	if err := g.datastore().Get(ctx, key, app); err != nil {
		return nil, err
	}

	ns := g.DeploymentName
	kc, err := getKubeClientset(ns)
	if err != nil {
		return app, nil
	}

	if app.Endpoint, err = sched.GetServiceEndpoint(kc, ns, name); err != nil {
		return app, nil
	}

	_, err = g.datastore().Put(ctx, key, app)
	return app, err
}

func (g *GCPCloud) AppDelete(name string) error {
	g.deleteAppFromCluster(name)
	return g.deleteAppFromDatastore(name)
}

func (g *GCPCloud) deleteAppFromDatastore(name string) error {
	store := g.datastore()
	ctx := g.ctxNS()

	q := datastore.NewQuery(buildKind).Filter("App =", name).KeysOnly()
	if err := deleteFromQuery(store, ctx, q); err != nil {
		return err
	}

	q = datastore.NewQuery(releaseKind).Filter("App =", name).KeysOnly()

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
	ns := g.DeploymentName
	kube, err := getKubeClientset(ns)
	if err != nil {
		return err
	}

	if _, err := kube.Core().Services(ns).Get(name); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	} else if err := kube.Core().Services(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
		return err
	}

	labels := klabels.Set(map[string]string{"name": name}).AsSelector()

	dp, err := kube.Extensions().Deployments(ns).Get(name)
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	}

	zerors := int32(0)
	dp.Spec.Replicas = &zerors

	if dp, err = kube.Extensions().Deployments(ns).Update(dp); err != nil {
		return err
	}

	sched.WaitUntilUpdated(kube, ns, name)

	if err = kube.Extensions().Deployments(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
		return err
	}

	// delete replicasets by label name=app
	res, err := kube.Extensions().ReplicaSets(ns).List(kapi.ListOptions{LabelSelector: labels.String()})
	if err != nil {
		return err
	}

	for _, rs := range res.Items {
		if err := kube.Extensions().ReplicaSets(ns).Delete(rs.Name, &kapi.DeleteOptions{}); err != nil {
			log.Warn(err)
		}
	}

	if _, err = kube.Extensions().Ingresses(ns).Get(name); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		}
	} else if err = kube.Extensions().Ingresses(ns).Delete(name, &kapi.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}
