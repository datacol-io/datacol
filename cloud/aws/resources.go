package aws

import (
	"fmt"
	pb "github.com/dinesh/datacol/api/models"
)

var notImplemented = fmt.Errorf("Not Implemented yet.")

func (a *AwsCloud) ResourceGet(name string) (*pb.Resource, error) {
	return nil, notImplemented
}

func (a *AwsCloud) ResourceDelete(name string) error {
	return notImplemented
}

func (a *AwsCloud) ResourceList() (pb.Resources, error) {
	return nil, notImplemented
}

func (a *AwsCloud) ResourceCreate(name, kind string, params map[string]string) (*pb.Resource, error) {
	return nil, notImplemented
}

func (a *AwsCloud) ResourceLink(app, name string) (*pb.Resource, error) {
	return nil, notImplemented
}

func (a *AwsCloud) ResourceUnlink(app, name string) (*pb.Resource, error) {
	return nil, notImplemented
}
