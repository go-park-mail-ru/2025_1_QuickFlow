package http

import (
	"encoding/json"
	"net/http"

	"quickflow/internal/delivery/forms"
)

// WriteJSONError sends JSON error response.
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(forms.ErrorForm{Error: message})
}
