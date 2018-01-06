package google

import (
	"io"

	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
)

func (g *GCPCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	return sched.LogStreamReq(g.kubeClient(), w, g.DeploymentName, app, opts)
}
