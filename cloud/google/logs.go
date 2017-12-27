package google

import (
	"io"

	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
)

func (g *GCPCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	ns := g.DeploymentName
	c, err := getKubeClientset(ns)
	if err != nil {
		return err
	}

	return sched.LogStreamReq(c, w, ns, app, opts)
}
