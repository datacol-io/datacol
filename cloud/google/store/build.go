package store

import (
	"context"
	"fmt"
	"sort"

	"cloud.google.com/go/datastore"
	pb "github.com/datacol-io/datacol/api/models"
)

func (g *DSBackend) BuildGet(app, id string) (*pb.Build, error) {
	var b pb.Build
	ctx, key := g.nestedKey(buildKind, id)
	if err := g.datastore().Get(ctx, key, &b); err != nil {
		return nil, fmt.Errorf("fetching build: %v", err)
	}

	return &b, nil
}

func (g *DSBackend) BuildDelete(app, id string) error {
	ctx, key := g.nestedKey(buildKind, id)
	return g.datastore().Delete(ctx, key)
}

func (g *DSBackend) BuildList(app string, limit int64) (pb.Builds, error) {
	q := datastore.NewQuery(buildKind).Namespace(g.DeploymentName).
		Filter("app = ", app)

		// Limit(limit).
		// Order("-" + "created_at") // FIXME:: need compound for created_at in desc order

	var builds pb.Builds
	_, err := g.datastore().GetAll(context.Background(), q, &builds)

	sort.Slice(builds, func(i, j int) bool {
		return builds[i].CreatedAt > builds[j].CreatedAt
	})

	size := int64(len(builds))

	if limit > 0 && size < limit {
		limit = size
	}

	return builds[0:limit], err
}

func (g *DSBackend) BuildCreate(app string, req *pb.CreateBuildOptions) (*pb.Build, error) {
	return nil, notImplemented
}

func (g *DSBackend) BuildSave(b *pb.Build) error {
	if b.Id == "" {
		b.Id = generateId("B", 5)
	}

	ctx, key := g.nestedKey(buildKind, b.Id)
	if _, err := g.datastore().Put(ctx, key, b); err != nil {
		return fmt.Errorf("saving build err: %v", err)
	}

	return nil
}
