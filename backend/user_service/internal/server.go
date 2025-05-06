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
	file_client "quickflow/shared/client/file_service"
	"quickflow/shared/interceptors"
	proto "quickflow/shared/proto/user_service"
	grpc2 "quickflow/user_service/internal/delivery/grpc"
	"quickflow/user_service/internal/delivery/interceptor"
	"quickflow/user_service/internal/repository/postgres"
	"quickflow/user_service/internal/repository/redis"
	"quickflow/user_service/internal/usecase"
	get_env "quickflow/utils/get-env"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", micro_addr.DefaultUserServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcConn, err := grpc.NewClient(
		get_env.GetServiceAddr(micro_addr.DefaultFileServiceAddrEnv, micro_addr.DefaultFileServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(micro_addr.MaxMessageSize)),
	)

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConn.Close()

	db, err := sql.Open("pgx", postgres_config.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	fileService := file_client.NewFileClient(grpcConn)
	userRepo := postgres.NewPostgresUserRepository(db)
	profileRepo := postgres.NewPostgresProfileRepository(db)
	redisRepo := redis.NewRedisSessionRepository()
	userUserCase := usecase.NewUserUseCase(userRepo, redisRepo, profileRepo)
	profileUseCase := usecase.NewProfileService(profileRepo, userRepo, fileService)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.ErrorInterceptor,
			interceptors.RequestIDServerInterceptor(),
		),
		grpc.MaxRecvMsgSize(micro_addr.MaxMessageSize),
		grpc.MaxSendMsgSize(micro_addr.MaxMessageSize))
	proto.RegisterUserServiceServer(server, grpc2.NewUserServiceServer(userUserCase))
	proto.RegisterProfileServiceServer(server, grpc2.NewProfileServiceServer(profileUseCase))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
