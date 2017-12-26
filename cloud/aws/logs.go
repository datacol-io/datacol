package aws

import (
	"io"

	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
)

func (a *AwsCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	ns := a.DeploymentName
	c := a.kubeClient()

	return sched.LogStreamReq(c, w, ns, app, opts)
}
