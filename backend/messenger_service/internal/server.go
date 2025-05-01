package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	postgres_config "quickflow/config/postgres"
	grpc2 "quickflow/messenger_service/internal/delivery/grpc"
	"quickflow/messenger_service/internal/delivery/grpc/interceptor"
	"quickflow/messenger_service/internal/repository/postgres"
	"quickflow/messenger_service/internal/usecase"
	"quickflow/messenger_service/utils/validation"
	"quickflow/shared/client/file_service"
	"quickflow/shared/client/user_service"
	"quickflow/shared/proto/messenger_service"
)

func main() {
	listener, err := net.Listen("tcp", ":8084")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcConnFileService, err := grpc.NewClient(
		"127.0.0.1:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}

	grpcConnUserService, err := grpc.NewClient(
		"127.0.0.1:8083",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	defer grpcConnFileService.Close()

	db, err := sql.Open("pgx", postgres_config.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	messageValidator := validation.NewMessageValidator()
	chatValidator := validation.NewChatValidator()

	fileService := file_service.NewFileClient(grpcConnFileService)
	profileService := userclient.NewProfileClient(grpcConnUserService)

	chatRepo := postgres.NewPostgresChatRepository(db)
	messageRepo := postgres.NewPostgresMessageRepository(db)

	messageUseCase := usecase.NewMessageService(messageRepo, fileService, chatRepo, messageValidator)
	chatUseCase := usecase.NewChatUseCase(chatRepo, fileService, profileService, messageRepo, chatValidator)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.ErrorInterceptor,
			interceptor.RequestIDUnaryInterceptor,
		),
	)

	log.Printf("Server is listening on %s", listener.Addr().String())
	proto.RegisterChatServiceServer(server, grpc2.NewChatServiceServer(chatUseCase))
	proto.RegisterMessageServiceServer(server, grpc2.NewMessageServiceServer(messageUseCase))
	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
