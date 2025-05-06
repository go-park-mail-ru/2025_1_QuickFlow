package interceptor

import (
	"context"
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	messenger_errors "quickflow/messenger_service/internal/errors"
)

func statusWithDetails(code codes.Code, msg, reason string) error {
	st := status.New(code, msg)

	details := &errdetails.ErrorInfo{
		Reason: reason,
		Domain: "messenger_service",
	}

	withDetails, err := st.WithDetails(details)
	if err != nil {
		return st.Err()
	}

	return withDetails.Err()
}

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

	if err == nil {
		return resp, nil
	}

	switch {
	case errors.Is(err, messenger_errors.ErrNotFound):
		return nil, statusWithDetails(codes.NotFound, err.Error(), "NOT_FOUND")

	case errors.Is(err, messenger_errors.ErrNotParticipant):
		return nil, statusWithDetails(codes.PermissionDenied, err.Error(), "NOT_PARTICIPANT")

	case errors.Is(err, messenger_errors.ErrInvalidChatCreationInfo):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "INVALID_CHAT_CREATION_INFO")

	case errors.Is(err, messenger_errors.ErrAlreadyInChat):
		return nil, statusWithDetails(codes.AlreadyExists, err.Error(), "ALREADY_IN_CHAT")

	case errors.Is(err, messenger_errors.ErrInvalidChatType):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "INVALID_CHAT_TYPE")

	case errors.Is(err, messenger_errors.ErrInvalidNumMessages):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "INVALID_NUM_MESSAGES")

	default:
		return nil, statusWithDetails(codes.Internal, err.Error(), "INTERNAL_ERROR")
	}
}
