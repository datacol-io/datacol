package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	pb "github.com/dinesh/datacol/api/controller"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"flag"
)

var (
	rpcPort     = 10000
	port        = 8080
	logPath string
	debugF  bool
)

func init(){
	flag.StringVar(&logPath, "log-dir", "", "path for logs")
	flag.BoolVar(&debugF, "debug", true, "debug mode")
}

func runRpcServer() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProviderServiceServer(grpcServer, newServer())

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
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

	fmt.Printf("http proxy on port: %d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func setupLogging(){
	flag.Parse()

	if len(logPath) > 0 {
		fmt.Printf("setting logging at %s\n", logPath)
		file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE, 0755)

		defer file.Close()
		if err != nil {
			panic(err)
		}
		log.SetOutput(file)
	}

	if debugF { log.SetLevel(log.DebugLevel) }
	
	log.SetFormatter(&log.TextFormatter{})
}

func main() {
	setupLogging()

	if err := runRpcServer(); err != nil {
		log.Fatal(err)
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
