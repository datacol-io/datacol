package local

import (
	"fmt"
	"io"
	"os"

	pb "github.com/datacol-io/datacol/api/models"
	sched "github.com/datacol-io/datacol/k8s"
)

func (g *LocalCloud) K8sConfigPath() (string, error) {
	return os.Getenv("HOME") + "/.kube/config", nil
}

func (g *LocalCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	return sched.LogStreamReq(g.kubeClient(), w, g.Name, app, opts)
}

func (g *LocalCloud) latestImage(app *pb.App) string {
	return fmt.Sprintf("%v/%v:%v", g.RegistryAddress, app.Name, app.BuildId)
}
