package google

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
)

func (g *GCPCloud) ReleaseGet(app, id string) (*pb.Release, error) {
	var r pb.Release
	ctx, key := g.nestedKey(releaseKind, id)
	if err := g.datastore().Get(ctx, key, &r); err != nil {
		return nil, fmt.Errorf("fetching release: %v", err)
	}

	return &r, nil
}

func (g *GCPCloud) ReleaseList(app string, limit int) (pb.Releases, error) {
	q := datastore.NewQuery(releaseKind).Namespace(g.DeploymentName).Filter("app = ", app).Limit(limit)

	var rs pb.Releases
	_, err := g.datastore().GetAll(context.Background(), q, &rs)

	return rs, err
}

func (g *GCPCloud) ReleaseDelete(app, id string) error {
	ctx, key := g.nestedKey(buildKind, id)
	return g.datastore().Delete(ctx, key)
}

func (g *GCPCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	r := &pb.Release{
		Id:        generateId("R", 5),
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
		Version:   g.releaseCount(b.App) + 1,
	}

	ctx, key := g.nestedKey(releaseKind, r.Id)
	_, err := g.datastore().Put(ctx, key, r)

	return r, err
}

func (g *GCPCloud) ReleasePromote(name, id string) error {
	r, err := g.ReleaseGet(name, id)
	if err != nil {
		return err
	}

	if r.BuildId == "" {
		return fmt.Errorf("No build found for release: %s", id)
	}

	b, err := g.BuildGet(name, r.BuildId)
	if err != nil {
		return err
	}

	image := fmt.Sprintf("gcr.io/%v/%v:%v", g.Project, b.App, b.Id)
	envVars, err := g.EnvironmentGet(b.App)
	if err != nil {
		return err
	}

	app, err := g.AppGet(name)
	if err != nil {
		return err
	}
	rversion := fmt.Sprintf("%d", r.Version)
	if err := common.UpdateApp(g.kubeClient(), b, g.DeploymentName,
		image, g.appLinkedDB(app),
		app.Domains, envVars,
		cloud.GCPProvider, rversion); err != nil {
		return err
	}

	app.ReleaseId = r.Id
	app.BuildId = b.Id

	return g.saveApp(app)
}

func (g *GCPCloud) releaseCount(app string) (version int64) {
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
