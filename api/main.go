package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/controller"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"time"
)

var (
	rpcPort = 10000
	port    = 8080
	logPath string
	debugF  bool
	timeout int
)

func init() {
	flag.StringVar(&logPath, "log-file", "", "path for logs")
	flag.BoolVar(&debugF, "debug", true, "debug mode")
	flag.IntVar(&timeout, "timeout", 2, "wait timeout for rpc proxy")
}

func runRpcServer() error {
	go func() {
		if err := newServer().Run(); err != nil {
			log.Fatal(fmt.Errorf("serveGRPC err: %v", err))
		}
	}()

	return nil
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := pb.RegisterProviderServiceHandlerFromEndpoint(ctx, gwmux, fmt.Sprintf("localhost:%d", rpcPort), opts); err != nil {
		return fmt.Errorf("serve: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	fmt.Printf("starting proxy on %d and grpc on %d ...\n", port, rpcPort)
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

	if err := runRpcServer(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(timeout) * time.Second)

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
