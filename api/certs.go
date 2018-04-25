package main

import (
	pbs "github.com/datacol-io/datacol/api/controller"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

func (s *Server) CertificateCreate(ctx context.Context, req *pbs.CreateCertReq) (*empty.Empty, error) {
	err := s.Provider.CertificateCreate(req.App, req.Domain, req.Crt, req.Key)
	return &empty.Empty{}, err
}
