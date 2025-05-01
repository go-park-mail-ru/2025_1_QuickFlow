package interceptor

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	file_errors "quickflow/file_service/internal/errors"
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
		case errors.Is(err, file_errors.ErrTooManyFiles) ||
			errors.Is(err, file_errors.ErrInvalidFileName) ||
			errors.Is(err, file_errors.ErrUnsupportedFileType) ||
			errors.Is(err, file_errors.ErrInvalidFileSize):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}
