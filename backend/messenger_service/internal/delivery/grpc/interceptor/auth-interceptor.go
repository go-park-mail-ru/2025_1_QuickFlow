package interceptor

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"quickflow/shared/models"
)

type AuthUseCase interface {
	LookupUserSession(ctx context.Context, session models.Session) (models.User, error)
}

func AuthInterceptor(authUseCase AuthUseCase) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("session metadata not found in context")
		}

		user_id := md.Get("user_id")
		if len(user_id) == 0 {
			return nil, errors.New("session user not found in context")
		}

		ctx = context.WithValue(ctx, "user_id", user_id[0])
		return handler(ctx, req)
	}
}
