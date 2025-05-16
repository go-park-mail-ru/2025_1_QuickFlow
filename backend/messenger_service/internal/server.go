package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	addr "quickflow/config/micro-addr"
	postgresConfig "quickflow/config/postgres"
	grpc2 "quickflow/messenger_service/internal/delivery/grpc"
	"quickflow/messenger_service/internal/delivery/grpc/interceptor"
	"quickflow/messenger_service/internal/repository/postgres"
	"quickflow/messenger_service/internal/usecase"
	"quickflow/messenger_service/utils/validation"
	"quickflow/metrics"
	"quickflow/shared/client/file_service"
	"quickflow/shared/client/user_service"
	"quickflow/shared/interceptors"
	"quickflow/shared/logger"
	"quickflow/shared/proto/messenger_service"
	getEnv "quickflow/utils/get-env"
)

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.DefaultMessengerServicePort))
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

	grpcConnUserService, err := grpc.NewClient(
		getEnv.GetServiceAddr(addr.DefaultUserServiceAddrEnv, addr.DefaultUserServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(addr.MaxMessageSize)),
	)

	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	defer grpcConnFileService.Close()

	db, err := sql.Open("pgx", postgresConfig.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	messageValidator := validation.NewMessageValidator()
	chatValidator := validation.NewChatValidator()

	fileService := file_service.NewFileClient(grpcConnFileService)
	profileService := userclient.NewProfileClient(grpcConnUserService)

	chatRepo := postgres.NewPostgresChatRepository(db)
	messageRepo := postgres.NewPostgresMessageRepository(db)

	messageUseCase := usecase.NewMessageService(messageRepo, fileService, chatRepo, messageValidator)
	chatUseCase := usecase.NewChatUseCase(chatRepo, fileService, profileService, messageRepo, chatValidator)

	messengerMetrics := metrics.NewMetrics("QuickFlow")

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.ErrorInterceptor,
			interceptor.RequestIDUnaryInterceptor,
			interceptors.RequestIDServerInterceptor(),
			interceptors.MetricsInterceptor(addr.DefaultMessengerServiceName, messengerMetrics),
		),
		grpc.MaxRecvMsgSize(addr.MaxMessageSize),
		grpc.MaxSendMsgSize(addr.MaxMessageSize),
	)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := addr.DefaultMessengerServicePort + 1000
		logger.Info(context.Background(), fmt.Sprintf("Metrics server is running on :%d/metrics", metricsPort))
		if err = http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			log.Fatalf("failed to start metrics HTTP server: %v", err)
		}
	}()

	log.Printf("Server is listening on %s", listener.Addr().String())
	proto.RegisterChatServiceServer(server, grpc2.NewChatServiceServer(chatUseCase))
	proto.RegisterMessageServiceServer(server, grpc2.NewMessageServiceServer(messageUseCase))
	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
