package main

import (
	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func internalError(err error, message string) error {
	log.Errorf(err.Error())
	return grpc.Errorf(codes.Unknown, message)
}
