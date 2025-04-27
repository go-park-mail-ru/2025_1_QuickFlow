package main

import (
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"

	minio_cofig "quickflow/file_service/config/minio"
	validation_config "quickflow/file_service/config/validation"
	grpc2 "quickflow/file_service/internal/delivery/grpc"
	"quickflow/file_service/internal/delivery/grpc/proto"
	"quickflow/file_service/internal/repository/minio"
	"quickflow/file_service/internal/usecase"
	"quickflow/file_service/utils/validation"
)

func main() {
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	minioPath := flag.String("minio-config", "../deploy/config/minio/config.toml", "Path to MinIO config file")
	validationPath := flag.String("validation-config", "../deploy/config/validation/config.toml", "Path to validation config file")
	flag.Parse()

	minioCfg, err := minio_cofig.ParseMinio(*minioPath)
	if err != nil {
		log.Fatalf("failed to parse minio config: %v", err)
	}
	validationConfig, err := validation_config.NewValidationConfig(*validationPath)
	if err != nil {
		log.Fatalf("failed to parse validation config: %v", err)
	}
	fileValidator := validation.NewFileValidator(validationConfig)

	fileRepo, err := minio.NewMinioRepository(minioCfg)
	if err != nil {
		log.Fatalf("failed to create minio repository: %v", err)
	}
	fileUseCase := usecase.NewFileUseCase(fileRepo, fileValidator)

	server := grpc.NewServer()
	proto.RegisterFileServiceServer(server, grpc2.NewFileServiceServer(fileUseCase))
	log.Printf("Server is listening on %s", listener.Addr().String())

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
