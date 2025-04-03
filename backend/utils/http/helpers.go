package http

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"quickflow/pkg/logger"

	"quickflow/internal/delivery/forms"
)

// WriteJSONError sends JSON error response.
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(forms.ErrorForm{Error: message})
}

func SetRequestId(ctx context.Context) context.Context {
	return context.WithValue(ctx,
		logger.RequestID,
		logger.ReqIdKey(uuid.New().String()))
}
