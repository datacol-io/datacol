package aws

import (
  pb "github.com/dinesh/datacol/api/models"
)

func (a *AwsCloud) AppList() (pb.Apps, error) {
  return nil, nil
}

func (a *AwsCloud) AppCreate(name string) (*pb.App, error) {
  return nil, nil
}

func (a *AwsCloud) AppRestart(app string) error {
  return nil
}

func (a *AwsCloud) AppGet(name string) (*pb.App, error) {
  return nil, nil
}

func (a *AwsCloud) AppDelete(name string) error {
  return nil
}



