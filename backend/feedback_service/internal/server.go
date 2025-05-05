package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	micro_addr "quickflow/config/micro-addr"
	postgres_config "quickflow/config/postgres"
	grpc3 "quickflow/feedback_service/internal/delivery/grpc"
	postgres2 "quickflow/feedback_service/internal/repository/postgres"
	"quickflow/feedback_service/internal/usecase"
	userclient "quickflow/shared/client/user_service"
	"quickflow/shared/interceptors"
	proto "quickflow/shared/proto/feedback_service"
	get_env "quickflow/utils/get-env"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", micro_addr.DefaultFeedbackServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcConnUserService, err := grpc.NewClient(
		get_env.GetServiceAddr(micro_addr.DefaultFileServiceAddrEnv, micro_addr.DefaultFileServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(micro_addr.MaxMessageSize)),
	)

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConnUserService.Close()

	db, err := sql.Open("pgx", postgres_config.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	profileService := userclient.NewProfileClient(grpcConnUserService)
	feedbackRepository := postgres2.NewFeedbackRepository(db)
	feedbackUseCase := usecase.NewFeedBackUseCase(feedbackRepository)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.RequestIDServerInterceptor()),
		grpc.MaxRecvMsgSize(micro_addr.MaxMessageSize),
		grpc.MaxSendMsgSize(micro_addr.MaxMessageSize))
	proto.RegisterFeedbackServiceServer(server, grpc3.NewFeedbackServiceServer(feedbackUseCase, profileService))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
