package main

import (
	log "github.com/Sirupsen/logrus"
	pbs "github.com/dinesh/datacol/api/controller"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func (s *Server) ProcessRun(srv pbs.ProviderService_ProcessRunServer) error {
	md, _ := metadata.FromIncomingContext(srv.Context())
	app, command := md["app"][0], md["command"][0]
	stream := &runStreamRW{srv}

	if err := s.Provider.ProcessRun(app, stream, command); err != nil {
		log.Errorf("failed run process: %v", err)
		return err
	}

	return nil
}

func (s *Server) ProcessList(ctx context.Context, req *pbs.AppRequest) (*pbs.ProcessListResponse, error) {
	items, err := s.Provider.ProcessList(req.Name)

	return &pbs.ProcessListResponse{Items: items}, err
}

type runStreamRW struct {
	stream pbs.ProviderService_ProcessRunServer
}

func (rs runStreamRW) Read(p []byte) (n int, err error) {
	msg, err := rs.stream.Recv()
	if err != nil {
		return len(p), err
	}

	copy(p, msg.Data)
	return len(p), nil
}

func (rs runStreamRW) Write(p []byte) (n int, err error) {
	if err = rs.stream.Send(&pbs.StreamMsg{
		Data: p,
	}); err != nil {
		log.Errorf("sending bytes to stream %s: %v", string(p), err)
	}

	return len(p), err
}
