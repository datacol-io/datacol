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

func (g *GCPCloud) ReleaseList(app string, limit int64) (pb.Releases, error) {
	q := datastore.NewQuery(releaseKind).Namespace(g.DeploymentName).Filter("app = ", app)

	if limit > 0 {
		q = q.Limit(int(limit))
	}

	var rs pb.Releases
	_, err := g.datastore().GetAll(context.Background(), q, &rs)

	return rs, err
}

func (g *GCPCloud) ReleaseDelete(app, id string) error {
	ctx, key := g.nestedKey(buildKind, id)
	return g.datastore().Delete(ctx, key)
}

func (g *GCPCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	image := fmt.Sprintf("gcr.io/%v/%v:%v", g.Project, b.App, b.Id)
	log.Debugf("---- Docker Image: %s", image)

	envVars, err := g.EnvironmentGet(b.App)
	if err != nil {
		return nil, err
	}

	app, err := g.AppGet(b.App)
	if err != nil {
		return nil, err
	}

	r := &pb.Release{
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
		Version:   g.store.ReleaseCount(b.App) + 1,
	}

	if err := g.store.ReleaseSave(r); err != nil {
		return r, err
	}

	rversion := fmt.Sprintf("%d", r.Version)

	if err := common.UpdateApp(g.kubeClient(), b, g.DeploymentName, image, g.appLinkedDB(app),
		app.Domains, envVars, cloud.GCPProvider, rversion); err != nil {
		return nil, err
	}

	app.ReleaseId = r.Id
	app.BuildId = b.Id

	return r, g.store.AppUpdate(app)
}
