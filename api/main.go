package main 

import (
  "golang.org/x/net/context"
  "net"
  "log"
  "fmt"
  pb "github.com/dinesh/datacol/api/models"
  "google.golang.org/grpc"
)

func main() {
  port := "8080"
  listener, err := net.Listen("tcp", ":" + port)
  if err != nil {
    log.Fatal(fmt.Errorf("failed to listen %v", err))
  }

  gs := grpc.NewServer()
  pb.RegisterStackServiceServer(gs, &StackServer{})
  fmt.Printf("Server is starting at %s ...\n", port)
  gs.Serve(listener)
}

type StackServer struct {

}

func (s *StackServer) Create(ctx context.Context, req *pb.CreateStackRequest) (*pb.Stack, error) {
  return &pb.Stack{}, nil
} 

