package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"io"
	"strings"
)

var (
	accessDeniedErr  = errors.New("Access denied.")
	emptyMetadataErr = errors.New("Unauthorized")
	apiKey           = "api_key"
	grpcAuth         = "grpcgateway-authorization"
)

func internalError(err error, message string) error {
	log.Errorf(err.Error())
	return grpc.Errorf(codes.Unknown, err.Error())
}

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}

func authorize(ctx context.Context, key string) bool {
	if md, ok := metadata.FromContext(ctx); ok {
		if len(md[apiKey]) > 0 && md[apiKey][0] == key {
			return true
		}

		if len(md[grpcAuth]) > 0 {
			return checkHttpAuthorization(md[grpcAuth][0], key)
		}

		return false
	}
	return false
}

func checkHttpAuthorization(value, expected string) bool {
	s := strings.SplitN(value, " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	return pair[1] == expected
}

func writeToFd(fd io.Writer, data []byte) error {
	w := 0
	n := len(data)
	for {
		nw, err := fd.Write(data[w:])
		if err != nil {
			return err
		}
		w += nw
		if nw >= n {
			return nil
		}
	}
}
