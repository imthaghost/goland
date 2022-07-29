package api

import (
	"context"
	pb "github.com/imthaghost/goland/zkp/rpc/zkp"
)

func (s *Server) HealthCheck(ctx context.Context, request *pb.HealthRequest) (*pb.HealthResponse, error) {

	resp := pb.HealthResponse{
		Status: int64(200),
	}

	return &resp, nil
}
