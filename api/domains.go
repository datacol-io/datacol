package main

import (
	"fmt"

	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

func (s *Server) AppUpdateDomain(ctx context.Context, req *pbs.AppResourceReq) (*empty.Empty, error) {
	empty := &empty.Empty{}
	if err := s.Provider.AppUpdateDomain(req.App, req.Resource); err != nil {
		return nil, internalError(err, "Failed to add domain")
	}

	// kick-off a new release to make it real into ingress

	app, err := s.Provider.AppGet(req.App)
	if err != nil {
		return nil, internalError(err, "Failed to fetch app")
	}

	if app.BuildId == "" {
		return empty, nil
	}

	b, err := s.Provider.BuildGet(app.Name, app.BuildId)
	if err != nil {
		return nil, internalError(err, fmt.Sprintf("Failed to fetch build by %s", app.BuildId))
	}

	_, err = s.Provider.BuildRelease(b, pb.ReleaseOptions{})

	return empty, err
}
