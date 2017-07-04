package aws

import (
  pb "github.com/dinesh/datacol/api/models"
)

func (a *AwsCloud) BuildGet(app, id string) (*pb.Build, error) {
  return nil, nil
}

func (a *AwsCloud) BuildDelete(app, id string) error {
  return nil
}

func (a *AwsCloud) BuildList(app string, limit int) (pb.Builds, error) {
  return nil, nil
}

func (g *AwsCloud) BuildCreate(app string, tarf []byte) (*pb.Build, error) {
  return nil, nil
}

func (a *AwsCloud) BuildImport(gskey string, tarf []byte) error {
  return nil
}

func (a *AwsCloud) BuildLogs(app, id string, index int) (int, []string, error) {
  return 0, []string{}, nil
}

func (a *AwsCloud) BuildRelease(b *pb.Build) (*pb.Release, error) {
  return nil, nil
}
