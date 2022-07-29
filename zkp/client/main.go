package main

import (
	"context"
	pb "github.com/imthaghost/goland/zkp/rpc/zkp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {

	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewAuthClient(conn)

	healthResp, err := client.HealthCheck(context.Background(), &pb.HealthRequest{})
	if err == nil {
		log.Println(healthResp)
	}

	registerResp, err := client.Register(
		context.Background(),
		&pb.RegisterRequest{
			Username: "imthaghost",
			Salt:     "982b332b832b3hjrv23rp39fb32fhk3823",
			GroupId:  "something",
		})

	if err == nil {
		log.Println(registerResp)
	}
}
