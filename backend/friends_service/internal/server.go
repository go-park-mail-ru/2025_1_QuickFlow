package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"

	"google.golang.org/grpc"

	micro_addr "quickflow/config/micro-addr"
	postgres_config "quickflow/config/postgres"
	grpc3 "quickflow/friends_service/internal/delivery/grpc"
	postgres "quickflow/friends_service/internal/repository"
	"quickflow/friends_service/internal/usecase"
	"quickflow/shared/interceptors"
	"quickflow/shared/proto/friends_service"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", micro_addr.DefaultFriendsServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	db, err := sql.Open("pgx", postgres_config.NewPostgresConfig().DatabaseFriendsUrl)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	friendsRepo := postgres.NewPostgresFriendsRepository(db)
	friendsUseCase := usecase.NewFriendsService(friendsRepo)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDServerInterceptor(),
		),
		grpc.MaxRecvMsgSize(micro_addr.MaxMessageSize),
		grpc.MaxSendMsgSize(micro_addr.MaxMessageSize),
	)

	log.Printf("Server is listening on %s", listener.Addr().String())
	proto.RegisterFriendsServiceServer(server, grpc3.NewFriendsServiceServer(friendsUseCase))
	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
