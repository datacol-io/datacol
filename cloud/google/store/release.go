package store

import (
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"cloud.google.com/go/datastore"
	pb "github.com/datacol-io/datacol/api/models"
)

func (g *DSBackend) ReleaseList(app string, limit int64) (pb.Releases, error) {
	q := datastore.NewQuery(releaseKind).Namespace(g.DeploymentName).Filter("app = ", app)
	if limit > 0 {
		q = q.Limit(int(limit))
	}

	var rs pb.Releases
	_, err := g.datastore().GetAll(context.Background(), q, &rs)

	return rs, err
}

func (g *DSBackend) ReleaseDelete(app, id string) error {
	ctx, key := g.nestedKey(releaseKind, id)
	return g.datastore().Delete(ctx, key)
}

func (g *DSBackend) ReleaseSave(b *pb.Release) error {
	ctx, key := g.nestedKey(releaseKind, b.Id)
	if _, err := g.datastore().Put(ctx, key, b); err != nil {
		return fmt.Errorf("saving release err: %v", err)
	}

	return nil
}

func (g *DSBackend) ReleaseCount(app string) (version int64) {
	q := datastore.NewQuery(releaseKind).
		Namespace(g.DeploymentName).
		Filter("app = ", app)

	total, err := g.datastore().Count(context.Background(), q)
	if err != nil {
		log.Warnf("fetching release total: %v", err)
	}

	version = int64(total)
	return
}
