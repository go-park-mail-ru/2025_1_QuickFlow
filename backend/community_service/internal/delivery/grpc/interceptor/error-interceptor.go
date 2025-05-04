package interceptor

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	community_errors "quickflow/community_service/internal/errors"
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
		case errors.Is(err, community_errors.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, community_errors.ErrorCommunityDescriptionTooLong) ||
			errors.Is(err, community_errors.ErrorCommunityNameTooShort) ||
			errors.Is(err, community_errors.ErrorCommunityNameTooLong) ||
			errors.Is(err, community_errors.ErrorCommunityAvatarSizeExceeded) ||
			errors.Is(err, community_errors.ErrorCommunityDescriptionTooLong):
			return status.Error(codes.InvalidArgument, err.Error()), nil
		case errors.Is(err, community_errors.ErrForbidden):
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case errors.Is(err, community_errors.ErrNotParticipant):
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case errors.Is(err, community_errors.ErrAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}
