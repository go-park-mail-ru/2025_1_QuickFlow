package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/utils"
)

type useCase interface {
	CreateUser(user models.User) (uuid.UUID, models.Session, error)
	GetUser(authData models.AuthForm) (models.Session, error)
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

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// validation
	if err := utils.Validate(user.Login, user.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// process data
	id, session, err := h.useCase.CreateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.SessionId.String(),
		Expires:  session.ExpireDate,
		HttpOnly: true,
		Secure:   true,
	})

	// return response
	body := map[string]interface{}{
		"user_id": id,
	}

	json.NewEncoder(w).Encode(&body)

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var authData models.AuthForm

	if err := json.NewDecoder(r.Body).Decode(&authData); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// process data
	session, err := h.useCase.GetUser(authData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.SessionId.String(),
		Expires:  session.ExpireDate,
		HttpOnly: true,
		Secure:   true,
	})

}
