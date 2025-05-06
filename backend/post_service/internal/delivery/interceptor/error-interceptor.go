package interceptor

import (
	"context"
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	post_errors "quickflow/post_service/internal/errors"
)

func withErrorInfo(code codes.Code, reason, message string) error {
	st := status.New(code, message)

	detail := &errdetails.ErrorInfo{
		Reason: reason, // e.g., "ERR_NOT_FOUND"
		Domain: "post_service",
	}

	stWithDetails, err := st.WithDetails(detail)
	if err != nil {
		return st.Err()
	}
	return stWithDetails.Err()
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
	case errors.Is(err, post_errors.ErrPostNotFound),
		errors.Is(err, post_errors.ErrNotFound):
		return nil, withErrorInfo(codes.NotFound, "ERR_NOT_FOUND", err.Error())

	case errors.Is(err, post_errors.ErrInvalidNumPosts),
		errors.Is(err, post_errors.ErrInvalidTimestamp),
		errors.Is(err, post_errors.ErrInvalidUUID):
		return nil, withErrorInfo(codes.InvalidArgument, "ERR_INVALID_ARGUMENT", err.Error())

	case errors.Is(err, post_errors.ErrPostDoesNotBelongToUser):
		return nil, withErrorInfo(codes.PermissionDenied, "ERR_FORBIDDEN", err.Error())

	case errors.Is(err, post_errors.ErrAlreadyExists):
		return nil, withErrorInfo(codes.AlreadyExists, "ERR_ALREADY_EXISTS", err.Error())

	case errors.Is(err, post_errors.ErrUploadFile):
		return nil, withErrorInfo(codes.Internal, "ERR_UPLOAD_FILE", err.Error())

	default:
		return nil, withErrorInfo(codes.Internal, "ERR_INTERNAL", err.Error())
	}
}
