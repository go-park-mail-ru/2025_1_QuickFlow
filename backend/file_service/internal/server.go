package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"

	addr "quickflow/config/micro-addr"
	postgresConfig "quickflow/config/postgres"
	minioConfig "quickflow/file_service/config/minio"
	validationConfig "quickflow/file_service/config/validation"
	grpc2 "quickflow/file_service/internal/delivery/grpc"
	"quickflow/file_service/internal/delivery/grpc/interceptor"
	"quickflow/file_service/internal/repository/minio"
	"quickflow/file_service/internal/repository/postgres"
	"quickflow/file_service/internal/usecase"
	"quickflow/file_service/utils/validation"
	"quickflow/metrics"
	"quickflow/shared/interceptors"
	"quickflow/shared/logger"
	proto "quickflow/shared/proto/file_service"
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
	// Конфиг-пути через флаги + resolve
	minioPathFlag := flag.String("minio-config", "", "Path to MinIO config file (relative)")
	validationPathFlag := flag.String("validation-config", "", "Path to validation config file (relative)")
	flag.Parse()

	db, err := sql.Open("pgx", postgresConfig.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	var minioPath, validationPath string
	if *minioPathFlag != "" {
		minioPath = resolveConfigPath(*minioPathFlag)
	} else {
		minioPath = resolveConfigPath("minio/config.toml")
	}

	if *validationPathFlag != "" {
		validationPath = resolveConfigPath(*validationPathFlag)
	} else {
		validationPath = resolveConfigPath("validation/config.toml")
	}

	// Чтение конфигов
	minioCfg, err := minioConfig.ParseMinio(minioPath)
	if err != nil {
		log.Fatalf("failed to parse minio config: %v", err)
	}
	validationCfg, err := validationConfig.NewValidationConfig(validationPath)
	if err != nil {
		log.Fatalf("failed to parse validation config: %v", err)
	}

	// Сервисы
	fileValidator := validation.NewFileValidator(validationCfg)
	fileStorage, err := minio.NewMinioRepository(minioCfg)
	if err != nil {
		log.Fatalf("failed to create minio repository: %v", err)
	}
	fileRepo := postgres.NewPostgresFileRepository(db)
	fileUseCase := usecase.NewFileUseCase(fileStorage, fileRepo, fileValidator)

	fileMetrics := metrics.NewMetrics("QuickFlow")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := addr.DefaultFileServicePort + 1000
		logger.Info(context.Background(), fmt.Sprintf("Metrics server is running on :%d/metrics", metricsPort))
		if err = http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			log.Fatalf("failed to start metrics HTTP server: %v", err)
		}
	}()

	// gRPC
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.DefaultFileServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptor.ErrorInterceptor,
		interceptors.RequestIDServerInterceptor(),
		interceptors.MetricsInterceptor(addr.DefaultFileServiceName, fileMetrics),
	),
		grpc.MaxRecvMsgSize(addr.MaxMessageSize),
		grpc.MaxSendMsgSize(addr.MaxMessageSize))
	proto.RegisterFileServiceServer(server, grpc2.NewFileServiceServer(fileUseCase))

	log.Printf("Server is listening on %s", listener.Addr().String())
	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
