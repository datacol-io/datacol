package local

import (
	"bufio"
	"io"
	"os"

	pb "github.com/dinesh/datacol/api/models"
	sched "github.com/dinesh/datacol/cloud/kube"
)

func (g *LocalCloud) K8sConfigPath() (string, error) {
	return os.Getenv("HOME") + "/.kube/config", nil
}

func (g *LocalCloud) LogStream(app string, opts pb.LogStreamOptions) (*bufio.Reader, func() error, error) {
	c, err := getKubeClientset(g.Name)
	if err != nil {
		return nil, nil, err
	}

	reader, err := sched.LogStreamReq(c, g.Name, app, opts)
	if err != nil {
		return nil, nil, err
	}

	return bufio.NewReader(reader), reader.Close, nil
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
