package interceptor

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	messenger_errors "quickflow/messenger_service/internal/errors"
)

func ErrorInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = status.Errorf(codes.Internal, "panic: %v", r)
		}
	}()

	resp, err = handler(ctx, req)

	if err != nil {

		switch {
		case errors.Is(err, messenger_errors.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, messenger_errors.ErrNotParticipant):
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case errors.Is(err, messenger_errors.ErrInvalidChatCreationInfo):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, messenger_errors.ErrAlreadyInChat):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case errors.Is(err, messenger_errors.ErrInvalidChatType):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, messenger_errors.ErrInvalidNumMessages):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}
