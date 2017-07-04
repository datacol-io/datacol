package aws

import (
  pb "github.com/dinesh/datacol/api/models"
)


func (a *AwsCloud) ReleaseList(app string, limit int) (pb.Releases, error) {
  return nil, nil
}

func (a *AwsCloud) ReleaseDelete(app, id string) error {
  return nil
}

