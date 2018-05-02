package local

import (
	"io"
	"io/ioutil"

	pb "github.com/datacol-io/datacol/api/models"
	"github.com/datacol-io/datacol/common"
)

func (g *LocalCloud) EnvironmentGet(name string) (pb.Environment, error) {
	return g.EnvMap[name], nil
}

func (g *LocalCloud) EnvironmentSet(name string, body io.Reader) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	g.EnvMap[name] = common.LoadEnvironment(data)
	return nil
}
