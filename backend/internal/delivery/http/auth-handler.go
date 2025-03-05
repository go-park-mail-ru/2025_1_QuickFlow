package http

import (
	"context"
	"encoding/json"
	"net/http"
	"quickflow/utils"

	"github.com/google/uuid"

	"quickflow/internal/models"
)

type AuthUseCase interface {
	CreateUser(user models.User) (uuid.UUID, models.Session, error)
	GetUser(authData models.AuthForm) (models.Session, error)
	LookupUserSession(ctx context.Context, session models.Session) (models.User, error)
}

type AuthHandler struct {
	authUseCase AuthUseCase
}

func NewAuthHandler(authUseCase AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Greet greets the user
func (a *AuthHandler) Greet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Hello World")
}

// SignUp creates new user.
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
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
	id, session, err := h.authUseCase.CreateUser(user)
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

// Login logs in user.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var authData models.AuthForm

	if err := json.NewDecoder(r.Body).Decode(&authData); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// process data
	session, err := h.authUseCase.GetUser(authData)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.SessionId.String(),
		Expires:  session.ExpireDate,
		HttpOnly: true,
		Secure:   true,
	})

	json.NewEncoder(w).Encode("залогинились")

	//http.Redirect(w, r, "/feed", http.StatusFound)
}
