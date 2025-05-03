package errors

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	st, ok := status.FromError(err)
	if !ok {
		// Не grpc-ошибка — внутренняя ошибка
		return New("INTERNAL", err.Error(), http.StatusInternalServerError)
	}

	switch st.Code() {
	case codes.NotFound:
		return New("NOT_FOUND", st.Message(), http.StatusNotFound)
	case codes.InvalidArgument:
		return New("INVALID_ARGUMENT", st.Message(), http.StatusBadRequest)
	case codes.PermissionDenied:
		return New("FORBIDDEN", st.Message(), http.StatusForbidden)
	case codes.Unauthenticated:
		return New("UNAUTHORIZED", st.Message(), http.StatusUnauthorized)
	case codes.AlreadyExists:
		return New("ALREADY_EXISTS", st.Message(), http.StatusConflict)
	case codes.DeadlineExceeded:
		return New("TIMEOUT", st.Message(), http.StatusGatewayTimeout)
	case codes.Unavailable:
		return New("SERVICE_UNAVAILABLE", st.Message(), http.StatusServiceUnavailable)
	default:
		return New("INTERNAL", st.Message(), http.StatusInternalServerError)
	}
}
