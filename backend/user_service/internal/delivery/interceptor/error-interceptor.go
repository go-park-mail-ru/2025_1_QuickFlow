package interceptor

import (
	"context"
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	user_errors "quickflow/user_service/internal/errors"
)

func withErrorInfo(code codes.Code, reason, message string) error {
	st := status.New(code, message)

	detail := &errdetails.ErrorInfo{
		Reason:   reason,
		Domain:   "user_service",
		Metadata: map[string]string{},
	}

	stWithDetails, err := st.WithDetails(detail)
	if err != nil {
		// fallback to status without details
		return st.Err()
	}
	return stWithDetails.Err()
}

func ErrorInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	switch {
	case errors.Is(err, user_errors.ErrNotFound),
		errors.Is(err, user_errors.ErrProfileNotFound),
		errors.Is(err, user_errors.ErrUserNotFound):
		return nil, withErrorInfo(codes.NotFound, "NOT_FOUND", err.Error())

	case errors.Is(err, user_errors.ErrAlreadyExists),
		errors.Is(err, user_errors.ErrUsernameTaken):
		return nil, withErrorInfo(codes.AlreadyExists, "ALREADY_EXISTS", err.Error())

	case errors.Is(err, user_errors.ErrInvalidUserId),
		errors.Is(err, user_errors.ErrInvalidProfileInfo),
		errors.Is(err, user_errors.ErrUserValidation),
		errors.Is(err, user_errors.ErrProfileValidation):
		return nil, withErrorInfo(codes.InvalidArgument, "INVALID_ARGUMENT", err.Error())

	default:
		return nil, withErrorInfo(codes.Internal, "INTERNAL", err.Error())
	}
}
