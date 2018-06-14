package store

import (
	pb "github.com/datacol-io/datacol/api/models"
)

type Store interface {
	AppCreate(*pb.App, *pb.AppCreateOptions) error
	AppUpdate(*pb.App) error
	AppGet(name string) (*pb.App, error)
	AppDelete(name string) error
	AppList() (pb.Apps, error)

	BuildList(app string, limit int64) (pb.Builds, error)
	BuildSave(*pb.Build) error
	BuildGet(app, id string) (*pb.Build, error)
	BuildDelete(app, id string) error

	ReleaseSave(*pb.Release) error
	ReleaseList(string, int64) (pb.Releases, error)
	ReleaseDelete(string, string) error
	ReleaseCount(string) int64
}
