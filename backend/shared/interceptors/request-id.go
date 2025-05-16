package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"quickflow/shared/logger"
)

func RequestIDClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		reqId, ok := ctx.Value(logger.RequestID).(logger.ReqIdKey)
		if ok && reqId != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "request-id", string(reqId))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func RequestIDServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if values := md.Get("request-id"); len(values) > 0 {
				ctx = context.WithValue(ctx, logger.RequestID, values[0])
			}
		}
		return handler(ctx, req)
	}
}
