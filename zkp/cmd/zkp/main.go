package main

import (
	"github.com/imthaghost/goland/zkp/internal/api"
	"github.com/imthaghost/goland/zkp/internal/store/inmemory"
	"log"
	"net"

	pb "github.com/imthaghost/goland/zkp/rpc/zkp"

	"google.golang.org/grpc"
)

func main() {

	ss := inmemory.New()

	lis, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	server := api.New(ss)
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterAuthServer(grpcServer, server)
	grpcServer.Serve(lis)
}
