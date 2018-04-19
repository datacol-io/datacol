package google

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/datastore"
	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/cloud/common"
	sched "github.com/datacol-io/datacol/cloud/kube"
)

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

	domains := app.Domains
	for _, domain := range strings.Split(options.Domain, ",") {
		domains = sched.MergeAppDomains(domains, domain)
	}

	r := &pb.Release{
		Id:        generateId("R", 5),
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
		Version:   g.releaseCount(b.App) + 1,
	}

	rversion := fmt.Sprintf("%d", r.Version)

	if err := common.UpdateApp(g.kubeClient(), b, g.DeploymentName, image, g.appLinkedDB(app),
		domains, envVars, cloud.GCPProvider, rversion); err != nil {
		return nil, err
	}

	if len(app.Domains) != len(domains) {
		app.Domains = domains

		if err = g.saveApp(app); err != nil {
			log.Warnf("datastore put failed: %v", err)
		}
	}

	ctx, key := g.nestedKey(releaseKind, r.Id)
	_, err = g.datastore().Put(ctx, key, r)

	if err != nil {
		return r, err
	}

	app.ReleaseId = r.Id
	app.BuildId = b.Id

	return r, g.saveApp(app)
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
