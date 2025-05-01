package main

import (
	"database/sql"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/jackc/pgx/v5/stdlib"

	postgres_config "quickflow/config/postgres"
	"quickflow/post_service/internal/client/file_sevice"
	"quickflow/post_service/internal/client/user_service"
	grpc3 "quickflow/post_service/internal/delivery/grpc"
	"quickflow/post_service/internal/repository/postgres"
	"quickflow/post_service/internal/usecase"
	"quickflow/post_service/utils/validation"
	"quickflow/shared/proto/post_service"
)

func main() {
	listener, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcConnFileService, err := grpc.Dial(
		"127.0.0.1:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	grpcConnUserService, err := grpc.Dial(
		"127.0.0.1:8083",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConnFileService.Close()

	db, err := sql.Open("pgx", postgres_config.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	fileService := file_sevice.NewFileClient(grpcConnFileService)
	postValidator := validation.NewPostValidator()
	postRepo := postgres.NewPostgresPostRepository(db)
	postUseCase := usecase.NewPostUseCase(postRepo, fileService, postValidator)
	userUseCase := user_service.NewUserClient(grpcConnUserService)

	server := grpc.NewServer()
	proto.RegisterPostServiceServer(server, grpc3.NewPostServiceServer(postUseCase, userUseCase))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
