package errors2

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	InternalErrorCode     = "INTERNAL"
	BadRequestErrorCode   = "BAD_REQUEST"
	UnauthorizedErrorCode = "UNAUTHORIZED"
)

type GatewayError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *GatewayError) Error() string {
	return fmt.Sprintf("gateway error: %s (%s)", e.Code, e.Message)
}

func New(code string, message string, status int) *GatewayError {
	return &GatewayError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

func FromGRPCError(err error) *GatewayError {
	if err == nil {
		return nil
	}

	// check if already a GatewayError
	var gwErr *GatewayError
	if errors.As(err, &gwErr) {
		return gwErr
	}

	st, ok := status.FromError(err)
	if !ok {
		// Не grpc-ошибка — внутренняя ошибка
		return New("INTERNAL", err.Error(), http.StatusInternalServerError)
	}

	// Попытка извлечь ErrorInfo
	code := "INTERNAL"
	for _, detail := range st.Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok {
			if info.Reason != "" {
				code = info.Reason
			}
			break
		}
	}

	return New(code, st.Message(), grpcCodeToHTTP(st.Code()))
}

func grpcCodeToHTTP(code codes.Code) int {
	switch code {
	case codes.NotFound:
		return http.StatusNotFound
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
