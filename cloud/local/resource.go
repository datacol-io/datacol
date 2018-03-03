package local

import (
	pb "github.com/datacol-io/datacol/api/models"
)

func (g *LocalCloud) ResourceGet(name string) (*pb.Resource, error) {
	return nil, nil
}

func (g *LocalCloud) ResourceDelete(name string) error {
	return nil
}

func (g *LocalCloud) ResourceList() (pb.Resources, error) {
	return nil, nil
}

func (g *LocalCloud) ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error) {
	return nil, nil
}

func (g *LocalCloud) ResourceLink(app, name string) (*pb.Resource, error) {
	return nil, nil
}

func (g *LocalCloud) ResourceUnlink(app, name string) (*pb.Resource, error) {
	return nil, nil
}
