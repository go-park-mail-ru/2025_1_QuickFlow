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
		Reason: reason, // e.g., "NOT_FOUND"
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

	// Если ошибка уже содержит ErrorInfo — просто пробрасываем её
	if st, ok := status.FromError(err); ok {
		for _, detail := range st.Details() {
			if _, ok := detail.(*errdetails.ErrorInfo); ok {
				// Пробрасываем error как есть
				return nil, err
			}
		}
	}

	switch {
	case errors.Is(err, post_errors.ErrPostNotFound),
		errors.Is(err, post_errors.ErrNotFound):
		return nil, withErrorInfo(codes.NotFound, "NOT_FOUND", err.Error())

	case errors.Is(err, post_errors.ErrInvalidNumPosts),
		errors.Is(err, post_errors.ErrInvalidTimestamp),
		errors.Is(err, post_errors.ErrInvalidUUID),
		errors.Is(err, post_errors.ErrInvalidNumComments):
		return nil, withErrorInfo(codes.InvalidArgument, "INVALID_ARGUMENT", err.Error())

	case errors.Is(err, post_errors.ErrDoesNotBelongToUser):
		return nil, withErrorInfo(codes.PermissionDenied, "FORBIDDEN", err.Error())

	case errors.Is(err, post_errors.ErrAlreadyExists):
		return nil, withErrorInfo(codes.AlreadyExists, "ALREADY_EXISTS", err.Error())

	case errors.Is(err, post_errors.ErrUploadFile):
		return nil, withErrorInfo(codes.Internal, "UPLOAD_FILE", err.Error())

	default:
		return nil, withErrorInfo(codes.Internal, "INTERNAL", err.Error())
	}
}
