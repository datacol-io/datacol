package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	pbs "github.com/datacol-io/datacol/api/controller"
	pb "github.com/datacol-io/datacol/api/models"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
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
	app, width, height := headers.Get("app"), headers.Get("width"), headers.Get("height")
	tty, detach := headers.Get("tty"), headers.Get("detach")

	if app == "" {
		return fmt.Errorf("Missing require param: app")
	}

	var options pb.ProcessRunOptions

	options.Tty, _ = strconv.ParseBool(tty)
	options.Detach, _ = strconv.ParseBool(detach)
	if width != "" {
		options.Width, _ = strconv.Atoi(width)
	}

	if height != "" {
		options.Height, _ = strconv.Atoi(height)
	}

	options.Entrypoint = strings.Split(headers.Get("command"), "#")
	_, err := s.Provider.ProcessRun(app, ws, options)
	return err
}

func (s *Server) ProcessRun(ctx context.Context, req *pbs.ProcessRunReq) (*pbs.ProcessRunResp, error) {
	app, command := req.Name, req.Command

	job, err := s.Provider.ProcessRun(app, nil, pb.ProcessRunOptions{
		Entrypoint: command,
		Tty:        false,
		Detach:     true,
	})
	if err != nil {
		log.Errorf("failed to run the process: %v", err)
		return nil, err
	}

	return &pbs.ProcessRunResp{Name: job}, nil
}

func (s *Server) ProcessSave(ctx context.Context, req *pb.Formation) (*empty.Empty, error) {
	if err := s.Provider.ProcessSave(req.App, req.Structure); err != nil {
		return nil, internalError(err, "Failed to run process")
	}

	return &empty.Empty{}, nil
}

func (s *Server) ProcessLimits(ctx context.Context, req *pb.ResourceLimits) (*empty.Empty, error) {
	if err := s.Provider.ProcessLimits(req.App, req.Proctype, req.Limits); err != nil {
		return nil, internalError(err, "Failed to set resource limits")
	}

	return &empty.Empty{}, nil
}

func (s *Server) ProcessList(ctx context.Context, req *pbs.AppRequest) (*pbs.ProcessListResponse, error) {
	items, err := s.Provider.ProcessList(req.Name)

	return &pbs.ProcessListResponse{Items: items}, err
}
