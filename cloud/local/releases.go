package local

import (
	"fmt"

	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/common"
)

func (g *LocalCloud) ReleaseList(app string, limit int) (pb.Releases, error) {
	return nil, nil
}

func (g *LocalCloud) ReleaseGet(app, id string) (*pb.Release, error) {
	for _, b := range g.Releases {
		if b.Id == id {
			return b, nil
		}
	}

	return nil, fmt.Errorf("release not found")
}

func (g *LocalCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	r := &pb.Release{
		Id:      common.GenerateId("R", 5),
		App:     b.App,
		BuildId: b.Id,
		Status:  pb.StatusCreated,
		Version: int64(len(g.Releases) + 1),
	}

	g.Releases = append(g.Releases, r)

	return r, nil
}

func (g *LocalCloud) ReleasePromote(name, id string) error {
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

	image := fmt.Sprintf("%v/%v:%v", g.RegistryAddress, name, b.Id)

	envVars, err := g.EnvironmentGet(b.App)
	if err != nil {
		return err
	}

	app, err := g.AppGet(b.App)
	if err != nil {
		return err
	}

	if err := common.UpdateApp(g.kubeClient(), b, g.Name, image, false,
		app.Domains, envVars, cloud.LocalProvider, fmt.Sprintf("%d", r.Version)); err != nil {
		return err
	}

	g.Releases = append(g.Releases, r)
	app.BuildId = b.Id
	app.ReleaseId = r.Id

	return g.saveApp(app)
}
