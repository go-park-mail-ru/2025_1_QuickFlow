package interceptor

import (
	"context"
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	feedback_errors "quickflow/feedback_service/internal/errors"
)

// helper: создает gRPC error с деталями
func statusWithDetails(code codes.Code, msg, reason string) error {
	st := status.New(code, msg)

	detail := &errdetails.ErrorInfo{
		Reason: reason,
		Domain: "feedback_service",
	}

	stWithDetails, err := st.WithDetails(detail)
	if err != nil {
		// если что-то пошло не так, fallback
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
	case errors.Is(err, feedback_errors.ErrNotFound):
		return nil, statusWithDetails(codes.NotFound, err.Error(), "NOT_FOUND")

	case errors.Is(err, feedback_errors.ErrRespondent):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "INVALID_RESPONDENT")

	case errors.Is(err, feedback_errors.ErrRating):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "INVALID_RATING")

	case errors.Is(err, feedback_errors.ErrTextTooLong):
		return nil, statusWithDetails(codes.InvalidArgument, err.Error(), "TEXT_TOO_LONG")

	default:
		return nil, statusWithDetails(codes.Internal, err.Error(), "INTERNAL")
	}
}
