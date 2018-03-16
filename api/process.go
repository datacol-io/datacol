package main

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc/metadata"
)

func (s *Server) ResourceProxy(ws *websocket.Conn) error {
	headers := ws.Request().Header
	host := headers.Get("remotehost")
	port := headers.Get("remoteport")

	if host == "" {
		return errors.New("Missing required header: remotehost")
	}

	if port == "" {
		return errors.New("Missing required header: remoteport")
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", host, port), 3*time.Second)

	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return errors.New("connection timeout out")
	}

	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(2)
	go copyAsync(ws, conn, &wg)
	go copyAsync(conn, ws, &wg)
	wg.Wait()

	return nil
}

func (s *Server) ProcessRunWs(ws *websocket.Conn) error {
	headers := ws.Request().Header
	app := headers.Get("app")

	if app == "" {
		return fmt.Errorf("Missing require param: app")
	}

	return s.Provider.ProcessRun(app, ws, headers.Get("command"))
}

func (s *Server) ProcessRun(srv pbs.ProviderService_ProcessRunServer) error {
	md, _ := metadata.FromIncomingContext(srv.Context())
	app, command := md["app"][0], md["command"][0]
	stream := &runStreamRW{srv}

	if err := s.Provider.ProcessRun(app, stream, command); err != nil {
		log.Errorf("failed to run the process: %v", err)
		return err
	}

	return nil
}

func (s *Server) ProcessSave(ctx context.Context, req *pb.Formation) (*empty.Empty, error) {
	if err := s.Provider.ProcessSave(req.App, req.Structure); err != nil {
		return nil, internalError(err, "Failed to run process")
	}

	return &empty.Empty{}, nil
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
