package main

import (
	pbs "github.com/datacol-io/datacol/api/controller"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

func (s *Server) CertificateCreate(ctx context.Context, req *pbs.CertificateReq) (*empty.Empty, error) {
	err := s.Provider.CertificateCreate(req.App, req.Domain, req.CertEncoded, req.KeyEncoded)
	return &empty.Empty{}, err
}

func (s *Server) CertificateDelete(ctx context.Context, req *pbs.CertificateReq) (*empty.Empty, error) {
	err := s.Provider.CertificateDelete(req.App, req.Domain)
	return &empty.Empty{}, err
}
