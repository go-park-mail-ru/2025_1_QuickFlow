package interceptors

import (
	"context"
	"quickflow/shared/logger"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quickflow/metrics"
)

func MetricsInterceptor(serviceName string, metrics *metrics.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		duration := time.Since(start)

		logger.Info(ctx, "In Interceptor")

		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
			metrics.ErrorCounter.WithLabelValues(serviceName, info.FullMethod, statusCode.String()).Inc()
		}

		metrics.Hits.WithLabelValues(serviceName, info.FullMethod, statusCode.String()).Inc()
		metrics.Timings.WithLabelValues(serviceName, info.FullMethod).Observe(duration.Seconds())

		return resp, err
	}
}
