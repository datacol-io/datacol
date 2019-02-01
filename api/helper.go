package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	rollbarAPI "github.com/stvp/rollbar"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"google.golang.org/grpc/metadata"
)

var (
	accessDeniedErr  = errors.New("Access denied.")
	emptyMetadataErr = errors.New("Unauthorized")
	apiKey           = "api_key"
	grpcAuth         = "grpcgateway-authorization"
)

func internalError(err error, message string) error {
	log.Error(err)
	return err
}

func toJson(object interface{}) string {
	dump, err := json.MarshalIndent(object, " ", "  ")
	if err != nil {
		log.Fatal(fmt.Errorf("dumping json: %v", err))
	}
	return string(dump)
}

func authorize(ctx context.Context, key string) bool {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
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

func checkAPIKey(r *http.Request) bool {
	if os.Getenv("DATACOL_API_KEY") == "" {
		return true
	}

	auth := r.Header.Get("Authorization")

	if auth == "" {
		return false
	}

	if !strings.HasPrefix(auth, "Basic ") {
		return false
	}

	c, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))

	if err != nil {
		return false
	}

	parts := strings.SplitN(string(c), ":", 2)

	if len(parts) != 2 || parts[1] != os.Getenv("DATACOL_API_KEY") {
		return false
	}

	return true
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

type websocketFunc func(*websocket.Conn) error

func ws(at string, handler websocketFunc) websocket.Handler {
	return websocket.Handler(func(ws *websocket.Conn) {
		if !checkAPIKey(ws.Request()) {
			ws.Write([]byte("Unable to authenticate. Invalid ApiKey."))
			return
		}

		err := handler(ws)

		if err != nil {
			log.Errorf("ws %s: %v", at, err)
			ws.Write([]byte(fmt.Sprintf("ERROR: %v\n", err)))
			return
		}
	})
}

func copyAsync(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	io.Copy(dst, src)
}

func handlePanicErr(err error, rollbarToken string) {
	if err == nil {
		return
	}

	log.Error(err)

	if rollbarToken == "" || os.Getenv("DATACOL_PROVIDER") == "local" {
		return
	}

	rollbarAPI.Platform = "datacol-controller"
	rollbarAPI.Token = rollbarToken

	fields := []*rollbarAPI.Field{
		{"version", os.Getenv("DATACOL_VERSION")},
		{"provider", os.Getenv("DATACOL_PROVIDER")},
		{"os", runtime.GOOS},
		{"arch", runtime.GOARCH},
	}

	rollbarAPI.Error("error", err, fields...)
	rollbarAPI.Wait()
}
