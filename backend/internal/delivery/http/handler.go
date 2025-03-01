package http

import (
	"net/http"
)

type useCase interface {
}

type Handler struct {
	useCase useCase
}

func NewHandler(useCase useCase) *Handler {
	return &Handler{useCase: useCase}
}

// Greet returns "Hello, world!".
//
// Use /hello request.
func (h *Handler) Greet(w http.ResponseWriter, _ *http.Request) {
	// ctx := r.Context()

	_, err := w.Write([]byte("Hello, world!\n"))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
