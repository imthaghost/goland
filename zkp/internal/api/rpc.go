package api

import (
	"github.com/imthaghost/goland/zkp/internal/store"
	pb "github.com/imthaghost/goland/zkp/rpc/zkp"
)

type Server struct {
	StoreService store.Service

	pb.UnimplementedAuthServer
}

// New ...
func New(ss store.Service) *Server {
	return &Server{
		StoreService: ss,
	}
}
