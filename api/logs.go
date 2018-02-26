package main

import (
	"github.com/appscode/go/log"
	pbs "github.com/dinesh/datacol/api/controller"
	pb "github.com/dinesh/datacol/api/models"
	"github.com/golang/protobuf/ptypes"
)

func (s *Server) LogStream(req *pbs.LogStreamReq, stream pbs.ProviderService_LogStreamServer) error {
	since, err := ptypes.Duration(req.Since)
	if err != nil {
		return err
	}

	return s.Provider.LogStream(req.Name, logStreamW{stream}, pb.LogStreamOptions{
		Follow:   req.Follow,
		Since:    since,
		Proctype: req.Proctype,
	})
}

type logStreamW struct {
	stream pbs.ProviderService_LogStreamServer
}

func (rs logStreamW) Write(p []byte) (n int, err error) {
	if err = rs.stream.Send(&pbs.StreamMsg{
		Data: p,
	}); err != nil {
		log.Errorf("sending bytes to stream %s: %v", string(p), err)
	}

	return len(p), err
}
