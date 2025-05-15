package interceptor

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/genproto/googleapis/rpc/errdetails"

	community_errors "quickflow/community_service/internal/errors"
)

func statusWithDetails(code codes.Code, msg, errorCode string) error {
	st := status.New(code, msg)

	detail := &errdetails.ErrorInfo{
		Reason: errorCode,
		Domain: "community_service",
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

	// Если ошибка уже содержит ErrorInfo — просто пробрасываем её
	if st, ok := status.FromError(err); ok {
		for _, detail := range st.Details() {
			if _, ok := detail.(*errdetails.ErrorInfo); ok {
				// Пробрасываем error как есть
				return nil, err
			}
		}
	}

	if err != nil {
		switch {
		case errors.Is(err, community_errors.ErrorCommunityNameTooShort):
			return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "COMMUNITY_NAME_TOO_SHORT")
		case errors.Is(err, community_errors.ErrorCommunityDescriptionTooLong):
			return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "COMMUNITY_DESCRIPTION_TOO_LONG")
		case errors.Is(err, community_errors.ErrAlreadyExists):
			return nil, statusWithDetails(codes.AlreadyExists, err.Error(), "COMMUNITY_ALREADY_EXISTS")
		case errors.Is(err, community_errors.ErrNotFound):
			return nil, statusWithDetails(codes.NotFound, err.Error(), "COMMUNITY_NOT_FOUND")
		case errors.Is(err, community_errors.ErrForbidden):
			return nil, statusWithDetails(codes.PermissionDenied, err.Error(), "FORBIDDEN")
		default:
			return nil, statusWithDetails(codes.Internal, err.Error(), "INTERNAL")
		}
	}

	return resp, nil
}
