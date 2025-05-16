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

	addr "quickflow/config/micro-addr"
	postgresConfig "quickflow/config/postgres"
	grpc3 "quickflow/friends_service/internal/delivery/grpc"
	postgres "quickflow/friends_service/internal/repository"
	"quickflow/friends_service/internal/usecase"
	"quickflow/metrics"
	"quickflow/shared/interceptors"
	"quickflow/shared/logger"
	"quickflow/shared/proto/friends_service"
)

const serviceName = "friends_service"

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", addr.DefaultFriendsServicePort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	db, err := sql.Open("pgx", postgresConfig.NewPostgresConfig().GetURL())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	friendsRepo := postgres.NewPostgresFriendsRepository(db)
	friendsUseCase := usecase.NewFriendsService(friendsRepo)

	friendsMetrics := metrics.NewMetrics("QuickFlow")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsPort := addr.DefaultFriendsServicePort + 1000
		logger.Info(context.Background(), fmt.Sprintf("Metrics server is running on :%d/metrics", metricsPort))
		if err = http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil); err != nil {
			log.Fatalf("failed to start metrics HTTP server: %v", err)
		}
	}()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDServerInterceptor(),
			interceptors.MetricsInterceptor(serviceName, friendsMetrics),
		),
		grpc.MaxRecvMsgSize(addr.MaxMessageSize),
		grpc.MaxSendMsgSize(addr.MaxMessageSize),
	)

	log.Printf("Server is listening on %s", listener.Addr().String())
	proto.RegisterFriendsServiceServer(server, grpc3.NewFriendsServiceServer(friendsUseCase))
	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
