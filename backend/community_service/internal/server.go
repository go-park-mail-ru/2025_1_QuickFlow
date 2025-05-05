package main

import (
    "database/sql"
    "fmt"
    "log"
    "net"
    "os"
    "path/filepath"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    _ "github.com/jackc/pgx/v5/stdlib"

    "quickflow/community_service/internal/client/file_sevice"
    grpc3 "quickflow/community_service/internal/delivery/grpc"
    "quickflow/community_service/internal/delivery/grpc/interceptor"
    "quickflow/community_service/internal/repository/postgres"
    "quickflow/community_service/internal/usecase"
    "quickflow/community_service/utils/validation"
    micro_addr "quickflow/config/micro-addr"
    postgres_config "quickflow/config/postgres"

    "quickflow/community_service/config"
    "quickflow/shared/interceptors"
    proto "quickflow/shared/proto/community_service"
    get_env "quickflow/utils/get-env"
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

    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", micro_addr.DefaultCommunityServicePort))
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    defer listener.Close()

    grpcConnFileService, err := grpc.NewClient(
        get_env.GetServiceAddr(micro_addr.DefaultFileServiceAddrEnv, micro_addr.DefaultFileServicePort),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
        grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(micro_addr.MaxMessageSize)),
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
    communityRepo := postgres.NewSqlCommunityRepository(db)
    communityUseCase := usecase.NewCommunityUseCase(communityRepo, fileService, validation.NewCommunityValidator(*cfg))

    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            interceptors.RequestIDServerInterceptor(),
            interceptor.ErrorInterceptor),
        grpc.MaxRecvMsgSize(micro_addr.MaxMessageSize),
        grpc.MaxSendMsgSize(micro_addr.MaxMessageSize),
    )
    proto.RegisterCommunityServiceServer(server, grpc3.NewCommunityServiceServer(communityUseCase))
    log.Printf("Server is listening on %s", listener.Addr().String())

    if err = server.Serve(listener); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
