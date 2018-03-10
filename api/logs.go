package main

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"golang.org/x/net/websocket"
)

func (s *Server) LogStream(req *pbs.LogStreamReq, stream pbs.ProviderService_LogStreamServer) error {
	return fmt.Errorf("Logging over GRPC has been deprecated. Please use use websocket based Path")
}

func (s *Server) LogStreamWs(ws *websocket.Conn) error {
	headers := ws.Request().Header
	app := headers.Get("app")

	if app == "" {
		return fmt.Errorf("Missing require param: app")
	}

	var (
		since  time.Duration
		follow bool
	)

	if raw := headers.Get("since"); raw != "" {
		duration, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		since = duration
	}

	if raw := headers.Get("follow"); raw != "" {
		follow = raw == "true"
	}

	return s.Provider.LogStream(app, ws, pb.LogStreamOptions{
		Since:    since,
		Follow:   follow,
		Proctype: headers.Get("process"),
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
