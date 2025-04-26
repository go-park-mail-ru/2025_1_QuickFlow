package server

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"

	"quickflow/feedback-microservice/feedback"
)

func main() {
	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatalln("cant listet port", err)
	}

	server := grpc.NewServer()

	feedback.RegisterFeedbackServiceServer(server, NewFeedbackManager())

	fmt.Println("starting server at :8081")
	server.Serve(lis)
}
