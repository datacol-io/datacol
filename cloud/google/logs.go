package google

import (
	"errors"
	"io"

	pb "github.com/datacol-io/datacol/api/models"
	sched "github.com/datacol-io/datacol/k8s"
)

func (g *GCPCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	return sched.LogStreamReq(g.kubeClient(), w, g.DeploymentName, app, opts)
}

func (g *GCPCloud) DockerCredsGet() (*pb.DockerCred, error) {
	return nil, errors.New("not implemented")
}
