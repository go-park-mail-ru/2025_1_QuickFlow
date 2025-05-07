package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"google.golang.org/grpc"

	micro_addr "quickflow/config/micro-addr"
	minio_config "quickflow/file_service/config/minio"
	validation_config "quickflow/file_service/config/validation"
	grpc2 "quickflow/file_service/internal/delivery/grpc"
	"quickflow/file_service/internal/delivery/grpc/interceptor"
	"quickflow/file_service/internal/repository/minio"
	"quickflow/file_service/internal/usecase"
	"quickflow/file_service/utils/validation"
	"quickflow/shared/interceptors"
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
	minioCfg, err := minio_config.ParseMinio(minioPath)
	if err != nil {
		log.Fatalf("failed to parse minio config: %v", err)
	}
	validationCfg, err := validation_config.NewValidationConfig(validationPath)
	if err != nil {
		log.Fatalf("failed to parse validation config: %v", err)
	}

	// Сервисы
	fileValidator := validation.NewFileValidator(validationCfg)
	fileRepo, err := minio.NewMinioRepository(minioCfg)
	if err != nil {
		log.Fatalf("failed to create minio repository: %v", err)
	}
	fileUseCase := usecase.NewFileUseCase(fileRepo, fileValidator)

	// gRPC
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", micro_addr.DefaultFileServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptor.ErrorInterceptor,
		interceptors.RequestIDServerInterceptor()),
		grpc.MaxRecvMsgSize(micro_addr.MaxMessageSize),
		grpc.MaxSendMsgSize(micro_addr.MaxMessageSize))
	proto.RegisterFileServiceServer(server, grpc2.NewFileServiceServer(fileUseCase))

	log.Printf("Server is listening on %s", listener.Addr().String())
	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
