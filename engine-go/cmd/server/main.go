package main

import (
	"log"
	"net"

	grpc "google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	log.Println("gRPC server running on :50051")

	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
