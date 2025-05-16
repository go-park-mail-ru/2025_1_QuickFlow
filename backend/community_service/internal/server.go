package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"quickflow/community_service/config"
	grpc3 "quickflow/community_service/internal/delivery/grpc"
	"quickflow/community_service/internal/delivery/grpc/interceptor"
	"quickflow/community_service/internal/repository/postgres"
	"quickflow/community_service/internal/usecase"
	"quickflow/community_service/utils/validation"
	addr "quickflow/config/micro-addr"
	postgresConfig "quickflow/config/postgres"
	"quickflow/metrics"
	fileService "quickflow/shared/client/file_service"
	"quickflow/shared/interceptors"
	"quickflow/shared/logger"
	proto "quickflow/shared/proto/community_service"
	getEnv "quickflow/utils/get-env"
)

func resolveConfigPath(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	if _, ok := os.LookupEnv("RUNNING_IN_CONTAINER"); ok {
		return filepath.Join("/config", rel)
	}
	return filepath.Join("../deploy/config", rel)
}

func main() {
	communityConfigPath := resolveConfigPath("community/config.toml")
	cfg, err := config.NewCommunityConfig(communityConfigPath)
	if err != nil {
		log.Fatalf("failed to load community config: %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.DefaultCommunityServicePort))
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

	if err != nil {
		log.Fatalf("failed to connect to file service: %v", err)
	}
	defer grpcConnFileService.Close()

	db, err := sql.Open("pgx", postgresConfig.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	fileService := fileService.NewFileClient(grpcConnFileService)
	communityRepo := postgres.NewSqlCommunityRepository(db)
	communityUseCase := usecase.NewCommunityUseCase(communityRepo, fileService, validation.NewCommunityValidator(*cfg))

	communityMetrics := metrics.NewMetrics("QuickFlow")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := addr.DefaultCommunityServicePort + 1000
		logger.Info(context.Background(), fmt.Sprintf("Metrics server is running on :%d/metrics", metricsPort))
		if err = http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			log.Fatalf("failed to start metrics HTTP server: %v", err)
		}
	}()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDServerInterceptor(),
			interceptor.ErrorInterceptor,
			interceptors.MetricsInterceptor(addr.DefaultCommunityServiceName, communityMetrics),
		),
		grpc.MaxRecvMsgSize(addr.MaxMessageSize),
		grpc.MaxSendMsgSize(addr.MaxMessageSize),
	)
	proto.RegisterCommunityServiceServer(server, grpc3.NewCommunityServiceServer(communityUseCase))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
