package main

import (
	log "github.com/Sirupsen/logrus"
	pbs "github.com/dinesh/datacol/api/controller"
	"google.golang.org/grpc/metadata"
)

func (s *Server) ProcessRun(srv pbs.ProviderService_ProcessRunServer) error {
	md, _ := metadata.FromIncomingContext(srv.Context())
	app, command := md["app"][0], md["command"][0]
	stream := &runStream{srv}

	if err := s.Provider.ProcessRun(app, stream, command); err != nil {
		log.Errorf("failed run process: %v", err)
		return err
	}
	return nil
}

type runStream struct {
	stream pbs.ProviderService_ProcessRunServer
}

func (rs *runStream) Read(p []byte) (n int, err error) {
	msg, err := rs.stream.Recv()

	if err != nil {
		log.Errorf("runStream.Read: %v", err)
		return len(p), err
	}

	copy(p, msg.Data)

	return len(p), nil
}

func (rs *runStream) Write(p []byte) (n int, err error) {
	if err = rs.stream.Send(&pbs.StreamMsg{
		Data: p,
	}); err != nil {
		log.Errorf("runStream.Write: %v", err)
	}

	return len(p), err
}
