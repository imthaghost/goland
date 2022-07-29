package api

import (
	"context"
	"errors"
	"github.com/imthaghost/goland/zkp/internal/store"
	"net/http"

	pb "github.com/imthaghost/goland/zkp/rpc/zkp"
)

func (s *Server) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// try to create a user from the request
	if request == nil {

		return &pb.RegisterResponse{
			Status: http.StatusBadRequest,
		}, errors.New("cannot have empty request")
	}
	u := &store.User{
		Username: request.Username,
		Salt:     request.Salt,
		GroupID:  request.GroupId,
	}
	err := s.StoreService.CreateUser(u)
	if err != nil {
		return &pb.RegisterResponse{
			Status: http.StatusInternalServerError,
		}, errors.New("could not create user")
	}

	resp := pb.RegisterResponse{
		Status: http.StatusAccepted,
	}
	return &resp, nil
}
