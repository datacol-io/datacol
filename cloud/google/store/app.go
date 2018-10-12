package store

import (
	"context"

	"cloud.google.com/go/datastore"
	pb "github.com/datacol-io/datacol/api/models"
)

const (
	appKind     = "App"
	buildKind   = "Build"
	releaseKind = "Release"
)

func (g *DSBackend) AppList() (pb.Apps, error) {
	var apps pb.Apps

	q := datastore.NewQuery(appKind).Namespace(g.DeploymentName)
	if _, err := g.datastore().GetAll(context.Background(), q, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}

func (g *DSBackend) AppCreate(app *pb.App, req *pb.AppCreateOptions) error {
	name := app.Name
	app.Status = pb.StatusCreated

	ctx, key := g.nestedKey(appKind, name)
	_, err := g.datastore().Put(ctx, key, app)

	return err
}

func (g *DSBackend) AppUpdate(app *pb.App) error {
	ctx, key := g.nestedKey(appKind, app.Name)
	_, err := g.datastore().Put(ctx, key, app)
	return err
}

func (g *DSBackend) AppGet(name string) (*pb.App, error) {
	app := new(pb.App)

	ctx, key := g.nestedKey(appKind, name)
	if err := g.datastore().Get(ctx, key, app); err != nil {
		return nil, err
	}

	return app, nil
}

func (g *DSBackend) AppDelete(name string) error {
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
