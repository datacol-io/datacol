package google

import (
	"fmt"
	"io"

	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
	"google.golang.org/api/googleapi"
)

func (g *GCPCloud) EnvironmentGet(name string) (pb.Environment, error) {
	gskey := fmt.Sprintf("%s.env", name)
	data, err := g.gsGet(g.BucketName, gskey)
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 404 {
			return pb.Environment{}, nil
		}
		return nil, err
	}

	return loadEnv(data), nil
}

func (g *GCPCloud) EnvironmentSet(name string, body io.Reader) error {
	gskey := fmt.Sprintf("%s.env", name)
	return g.gsPut(g.BucketName, gskey, body)
}

func (g *GCPCloud) GetRunningPods(app string) (string, error) {
	ns := g.DeploymentName
	c, err := getKubeClientset(ns)
	if err != nil {
		return "", err
	}

	return sched.RunningPods(ns, app, c)
}
