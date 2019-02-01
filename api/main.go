package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/datacol-io/datacol/api/controller"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	rpcPort  = 10000
	port     = 8888
	logPath  string
	debugF   bool
	timeout  int
	crashing bool
	rbToken  string
)

func init() {
	flag.StringVar(&logPath, "log-file", "", "path for logs")
	flag.BoolVar(&debugF, "debug", true, "debug mode")
	flag.IntVar(&timeout, "timeout", 2, "wait timeout for rpc proxy")
}

func runRpcServer(server *Server) {
	go func() {
		if err := server.Run(); err != nil {
			log.Fatalf("serveGRPC err: %v", err)
		}
	}()
}

func runHttpServer(server *Server) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := pb.RegisterProviderServiceHandlerFromEndpoint(ctx, gwmux, fmt.Sprintf("localhost:%d", rpcPort), opts); err != nil {
		return fmt.Errorf("serve: %v", err)
	}

	mux := http.NewServeMux()

	// Implementing bidi-streaming using GRPC is a lot of headache,
	// it's better to use websockets protocols for streaming logs and interactive one-off commands
	mux.Handle("/ws/v1/logs", ws("appLogs", server.LogStreamWs))
	mux.Handle("/ws/v1/exec", ws("processExec", server.ProcessRunWs))
	mux.Handle("/ws/v1/proxy", ws("appProxy", server.ResourceProxy))
	mux.Handle("/ws/v1/builds/logs", ws("buildLogs", server.BuildLogStreamReq))
	mux.Handle("/ws/v1/builds/import", ws("buildImport", server.BuildImport))
	mux.Handle("/", gwmux)

	fmt.Printf("Starting server on http=%d and grpc=%d ports\n", port, rpcPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func setupLogging() {
	flag.Parse()

	if len(logPath) > 0 {
		fmt.Printf("setting logging at %s\n", logPath)
		file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE, 0755)

		if err != nil {
			panic(err)
		}
		log.SetOutput(file)
	}

	if debugF {
		log.SetLevel(log.DebugLevel)
	}

	log.SetFormatter(&log.TextFormatter{})
}

func main() {
	setupLogging()
	server := newServer()
	runRpcServer(server)
	time.Sleep(time.Duration(timeout) * time.Second)

	if rbToken == "" {
		log.Warnf("Won't send notifications to rollbar.")
	}

	if err := runHttpServer(server); err != nil {
		handlePanicErr(err, rbToken)
	}
}
