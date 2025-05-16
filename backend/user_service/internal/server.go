package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	addr "quickflow/config/micro-addr"
	postgresConfig "quickflow/config/postgres"
	"quickflow/metrics"
	fileClient "quickflow/shared/client/file_service"
	"quickflow/shared/interceptors"
	"quickflow/shared/logger"
	proto "quickflow/shared/proto/user_service"
	grpc2 "quickflow/user_service/internal/delivery/grpc"
	"quickflow/user_service/internal/delivery/interceptor"
	"quickflow/user_service/internal/repository/postgres"
	"quickflow/user_service/internal/repository/redis"
	"quickflow/user_service/internal/usecase"
	getEnv "quickflow/utils/get-env"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.DefaultUserServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcConn, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultFileServiceAddrEnv, addr.DefaultFileServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConn.Close()

	db, err := sql.Open("pgx", postgresConfig.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	fileService := fileClient.NewFileClient(grpcConn)
	userRepo := postgres.NewPostgresUserRepository(db)
	profileRepo := postgres.NewPostgresProfileRepository(db)
	redisRepo := redis.NewRedisSessionRepository()
	userUserCase := usecase.NewUserUseCase(userRepo, redisRepo, profileRepo)
	profileUseCase := usecase.NewProfileService(profileRepo, userRepo, fileService)

	userMetrics := metrics.NewMetrics("QuickFlow")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := addr.DefaultUserServicePort + 1000
		logger.Info(context.Background(), fmt.Sprintf("Metrics server is running on :%d/metrics", metricsPort))
		if err = http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			log.Fatalf("failed to start metrics HTTP server: %v", err)
		}
	}()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.ErrorInterceptor,
			interceptors.RequestIDServerInterceptor(),
			interceptors.MetricsInterceptor(addr.DefaultUserServiceName, userMetrics),
		),
		grpc.MaxRecvMsgSize(addr.MaxMessageSize),
		grpc.MaxSendMsgSize(addr.MaxMessageSize))
	proto.RegisterUserServiceServer(server, grpc2.NewUserServiceServer(userUserCase))
	proto.RegisterProfileServiceServer(server, grpc2.NewProfileServiceServer(profileUseCase))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
