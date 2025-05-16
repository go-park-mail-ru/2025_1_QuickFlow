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
	"quickflow/post_service/internal/client/file_sevice"
	"quickflow/post_service/internal/client/user_service"
	grpc3 "quickflow/post_service/internal/delivery/grpc"
	"quickflow/post_service/internal/repository/postgres"
	"quickflow/post_service/internal/usecase"
	"quickflow/post_service/utils/validation"
	"quickflow/shared/interceptors"
	"quickflow/shared/logger"
	"quickflow/shared/proto/post_service"
	getEnv "quickflow/utils/get-env"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.DefaultPostServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	grpcConnFileService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultFileServiceAddrEnv, addr.DefaultFileServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	grpcConnUserService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultUserServiceAddrEnv, addr.DefaultUserServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConnFileService.Close()

	db, err := sql.Open("pgx", postgresConfig.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	fileService := file_sevice.NewFileClient(grpcConnFileService)
	postValidator := validation.NewPostValidator()
	postRepo := postgres.NewPostgresPostRepository(db)
	postUseCase := usecase.NewPostUseCase(postRepo, fileService, postValidator)
	userUseCase := user_service.NewUserClient(grpcConnUserService)

	postMetrics := metrics.NewMetrics("QuickFlow")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := addr.DefaultPostServicePort + 1000
		logger.Info(context.Background(), fmt.Sprintf("Metrics server is running on :%d/metrics", metricsPort))
		if err = http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			log.Fatalf("failed to start metrics HTTP server: %v", err)
		}
	}()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDServerInterceptor(),
			interceptors.MetricsInterceptor(addr.DefaultPostServiceName, postMetrics),
		),
		grpc.MaxRecvMsgSize(addr.MaxMessageSize),
		grpc.MaxSendMsgSize(addr.MaxMessageSize))
	proto.RegisterPostServiceServer(server, grpc3.NewPostServiceServer(postUseCase, userUseCase))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
