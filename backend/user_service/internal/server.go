package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	postgres_config "quickflow/config/postgres"
	"quickflow/user_service/internal/client/file_sevice"
	grpc2 "quickflow/user_service/internal/delivery/grpc"
	"quickflow/user_service/internal/delivery/grpc/proto"
	"quickflow/user_service/internal/repository/postgres"
	"quickflow/user_service/internal/repository/redis"
	"quickflow/user_service/internal/usecase"
)

func main() {
	listener, err := net.Listen("tcp", ":8083")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcConn, err := grpc.NewClient(
		"localhost:8081",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConn.Close()

	db, err := sql.Open("pgx", postgres_config.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	fileService := file_sevice.NewFileClient(grpcConn)
	userRepo := postgres.NewPostgresUserRepository(db)
	profileRepo := postgres.NewPostgresProfileRepository(db)
	redisRepo := redis.NewRedisSessionRepository()
	userUserCase := usecase.NewAuthService(userRepo, redisRepo, profileRepo)
	profileUseCase := usecase.NewProfileService(profileRepo, userRepo, fileService)

	server := grpc.NewServer()
	proto.RegisterUserServiceServer(server, grpc2.NewUserServiceServer(userUserCase))
	proto.RegisterProfileServiceServer(server, grpc2.NewProfileServiceServer(profileUseCase))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
