package aws

import (
	"io"

	pb "github.com/datacol-io/datacol/api/models"
	sched "github.com/datacol-io/datacol/k8s"
)

func (a *AwsCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	return sched.LogStreamReq(a.kubeClient(), w, a.DeploymentName, app, opts)
}
