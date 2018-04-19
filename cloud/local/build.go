package local

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/crypto/rand"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/cloud"
	"github.com/datacol-io/datacol/cloud/common"
	docker "github.com/fsouza/go-dockerclient"
)

func (g *LocalCloud) BuildGet(app, id string) (*pb.Build, error) {
	for _, b := range g.Builds {
		if b.Id == id {
			return b, nil
		}
	}

	return nil, fmt.Errorf("build not found")
}

func (g *LocalCloud) BuildDelete(app, id string) error {
	return nil
}

func (g *LocalCloud) BuildList(app string, limit int) (pb.Builds, error) {
	return g.Builds, nil
}

func (g *LocalCloud) ReleaseList(app string, limit int) (pb.Releases, error) {
	return nil, nil
}

func (g *LocalCloud) ReleaseDelete(app, id string) error {
	return nil
}

func (g *LocalCloud) BuildCreate(app string, req *pb.CreateBuildOptions) (*pb.Build, error) {
	build := &pb.Build{
		App:      app,
		Id:       rand.Characters(5),
		Status:   "CREATED",
		Procfile: req.Procfile,
	}

	g.Builds = append(g.Builds, build)
	return build, nil
}

func (g *LocalCloud) BuildImport(id, filename string) error {
	build, err := g.BuildGet("", id)
	if err != nil {
		return err
	}

	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	dkr, app, id := dockerClient(), build.App, build.Id

	if err := dkr.BuildImage(docker.BuildImageOptions{
		Name:         app,
		InputStream:  r,
		OutputStream: os.Stdout,
	}); err != nil {
		return fmt.Errorf("failed to build image: %v", err)
	}

	repo := fmt.Sprintf("%s/%s", g.RegistryAddress, app)
	tag := id

	log.Debugf("Tagging image %s to %s:%s", app, repo, tag)

	if err := dkr.TagImage(app, docker.TagImageOptions{Repo: repo, Tag: tag}); err != nil {
		return fmt.Errorf("failed to tag image with %v: %v", tag, err)
	}

	log.Debugf("Pushing image %s:%s to %s", repo, tag, g.RegistryAddress)

	if err := dkr.PushImage(docker.PushImageOptions{
		Name:              repo,
		Tag:               tag,
		OutputStream:      os.Stdout,
		InactivityTimeout: 2 * time.Minute,
	}, docker.AuthConfiguration{
		ServerAddress: g.RegistryAddress,
	}); err != nil {
		return fmt.Errorf("failed to push image: %v", err)
	}

	return nil
}

func (g *LocalCloud) BuildLogs(app, id string, index int) (int, []string, error) {
	return 0, nil, nil
}

func (g *LocalCloud) BuildLogsStream(id string) (io.Reader, error) {
	return nil, nil
}

func (g *LocalCloud) BuildRelease(b *pb.Build, options pb.ReleaseOptions) (*pb.Release, error) {
	image := fmt.Sprintf("%v/%v:%v", g.RegistryAddress, b.App, b.Id)
	log.Printf("---- Docker Image: %s\n", image)

	envVars, err := g.EnvironmentGet(b.App)
	if err != nil {
		return nil, err
	}

	app, err := g.AppGet(b.App)
	if err != nil {
		return nil, err
	}

	r := &pb.Release{
		Id:      common.GenerateId("R", 5),
		App:     b.App,
		BuildId: b.Id,
		Status:  pb.StatusCreated,
		Version: int64(len(g.Releases) + 1),
	}

	if err := common.UpdateApp(g.kubeClient(), b, g.Name, image, false,
		[]string{}, envVars, cloud.LocalProvider, fmt.Sprintf("%d", r.Version)); err != nil {
		return nil, err
	}

	g.Releases = append(g.Releases, r)
	app.BuildId = b.Id
	app.ReleaseId = r.Id

	return r, g.saveApp(app)
}
