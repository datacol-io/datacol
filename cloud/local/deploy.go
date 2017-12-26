package local

import (
	"io"
	"os"

	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
)

func (g *LocalCloud) K8sConfigPath() (string, error) {
	return os.Getenv("HOME") + "/.kube/config", nil
}

func (g *LocalCloud) LogStream(app string, w io.Writer, opts pb.LogStreamOptions) error {
	c, err := getKubeClientset(g.Name)
	if err != nil {
		return err
	}

	return sched.LogStreamReq(c, w, g.Name, app, opts)
}

func (g *LocalCloud) ProcessRun(app string, stream io.ReadWriter, command string) error {
	ns := g.Name
	cfg, err := getKubeClientConfig(ns)
	if err != nil {
		return err
	}

	c, err := getKubeClientset(ns)
	if err != nil {
		return err
	}

	return sched.ProcessExec(c, cfg, ns, app, command, stream)
}
