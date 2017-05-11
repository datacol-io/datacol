package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func internalError(err error, message string) error {
	log.Errorf(err.Error())
	return grpc.Errorf(codes.Unknown, message)
}

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}
