package aws

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
)

func (a *AwsCloud) dynamoReleases() string {
	return fmt.Sprintf("%s-releases", a.DeploymentName)
}

func (a *AwsCloud) ReleaseList(app string, limit int64) (pb.Releases, error) {
	return a.store.ReleaseList(app, limit)
}

func (a *AwsCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	image := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s",
		os.Getenv("AWS_ACCOUNT_ID"), a.Region, a.ecrRepository(b.App), b.Id,
	)
	log.Debugf("---- Docker Image: %s", image)

	app, err := a.AppGet(b.App)
	if err != nil {
		return nil, err
	}

	envVars, err := a.EnvironmentGet(b.App)
	if err != nil {
		return nil, err
	}

	r := &pb.Release{
		Id:        generateId("R", 5),
		App:       b.App,
		BuildId:   b.Id,
		Status:    pb.StatusCreated,
		CreatedAt: timestampNow(),
		Version:   a.store.ReleaseCount(b.App) + 1,
	}

	if err = a.store.ReleaseSave(r); err != nil {
		return r, err
	}

	app.BuildId = b.Id
	app.ReleaseId = r.Id
	rversion := fmt.Sprintf("%d", r.Version)

	log.Debugf("Saving app state: %s err:%v", toJson(app), a.store.AppUpdate(app)) // note the mutate function

	if err := common.UpdateApp(a.kubeClient(), b, a.DeploymentName,
		image, false, app.Domains, envVars, cloud.AwsProvider, rversion); err != nil {
		return nil, err
	}

	//TODO: update release status

	return r, nil
}

func (a *AwsCloud) ReleaseDelete(app, id string) error {
	return a.store.ReleaseDelete(app, id)
}
