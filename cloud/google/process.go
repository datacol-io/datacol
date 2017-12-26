package google

import (
	"io"

	"github.com/dinesh/datacol/cloud/kube"
)

func (g *GCPCloud) ProcessRun(app string, stream io.ReadWriter, command string) error {
	ns := g.DeploymentName
	cfg, err := getKubeClientConfig(ns)
	if err != nil {
		return err
	}

	c, err := getKubeClientset(ns)
	if err != nil {
		return err
	}

	return kube.ProcessExec(c, cfg, ns, app, command, stream)
}
