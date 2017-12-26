package local

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/appscode/go/crypto/rand"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/dinesh/datacol/cloud/common"
	sched "github.com/dinesh/datacol/cloud/kube"
	docker "github.com/fsouza/go-dockerclient"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	return nil, fmt.Errorf("not implemented.")
}

func (g *LocalCloud) BuildImport(app, filename string) (*pb.Build, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	dkr, id := dockerClient(), rand.Characters(5)

	if err := dkr.BuildImage(docker.BuildImageOptions{
		Name:         app,
		InputStream:  r,
		OutputStream: os.Stdout,
	}); err != nil {
		return nil, fmt.Errorf("failed to build image: %v", err)
	}

	repo := fmt.Sprintf("%s/%s", g.RegistryAddress, app)
	tag := id

	log.Debugf("Tagging image %s to %s", app, repo, tag)

	if err := dkr.TagImage(app, docker.TagImageOptions{Repo: repo, Tag: tag}); err != nil {
		return nil, fmt.Errorf("failed to tag image with %v: %v", tag, err)
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
		return nil, fmt.Errorf("failed to push image: %v", err)
	}

	build := &pb.Build{
		App:    app,
		Id:     id,
		Status: "CREATED",
	}

	g.Builds = append(g.Builds, build)
	return build, nil
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

	c, err := getKubeClientset(g.Name)
	if err != nil {
		return nil, err
	}

	deployer, err := sched.NewDeployer(c)
	if err != nil {
		return nil, err
	}

	port := 8080
	if pv, ok := envVars["PORT"]; ok {
		p, err := strconv.Atoi(pv)
		if err != nil {
			return nil, err
		}
		port = p
	}

	ret, err := deployer.Run(&sched.DeployRequest{
		ServiceID:     b.App,
		Image:         image,
		Replicas:      1,
		Environment:   g.Name,
		ContainerPort: intstr.FromInt(port),
		EnvVars:       envVars,
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Deployed %s with %+v\n", b.App, ret.Request)

	r := &pb.Release{
		Id:      common.GenerateId("R", 5),
		App:     b.App,
		BuildId: b.Id,
		Status:  pb.StatusCreated,
	}

	g.Releases = append(g.Releases, r)

	return r, err
}