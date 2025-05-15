package interceptor

import (
	"context"
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	file_errors "quickflow/file_service/internal/errors"
)

func statusWithDetails(code codes.Code, msg, reason string) error {
	st := status.New(code, msg)

	detail := &errdetails.ErrorInfo{
		Reason: reason,
		Domain: "file_service",
	}

	stWithDetails, err := st.WithDetails(detail)
	if err != nil {
		// fallback if details fail
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
	case errors.Is(err, file_errors.ErrTooManyFiles):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "TOO_MANY_FILES")

	case errors.Is(err, file_errors.ErrInvalidFileName):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "INVALID_FILE_NAME")

	case errors.Is(err, file_errors.ErrUnsupportedFileType):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "UNSUPPORTED_FILE_TYPE")

	case errors.Is(err, file_errors.ErrInvalidFileSize):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "INVALID_FILE_SIZE")

	default:
		return nil, statusWithDetails(codes.Internal, err.Error(), "INTERNAL")
	}
}
