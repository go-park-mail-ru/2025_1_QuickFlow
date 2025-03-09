package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"

	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	"quickflow/utils"
)

type AuthUseCase interface {
	CreateUser(ctx context.Context, user models.User) (uuid.UUID, models.Session, error)
	GetUser(ctx context.Context, authData models.LoginData) (models.Session, error)
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
func (a *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var form forms.SignUpForm

	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// converting transport form to domain model
	user := models.User{
		Login:       form.Login,
		Name:        form.Name,
		Surname:     form.Surname,
		Sex:         form.Sex,
		DateOfBirth: form.DateOfBirth,
		Password:    form.Password,
	}

	// validation
	if err := utils.Validate(user.Login, user.Password, user.Name, user.Surname); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// process data
	id, session, err := a.authUseCase.CreateUser(r.Context(), user)
	if err != nil {
		log.Println(err.Error())
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
func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var form forms.AuthForm

	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// converting transport form to domain model
	loginData := models.LoginData{
		Login:    form.Login,
		Password: form.Password,
	}

	// process data
	session, err := a.authUseCase.GetUser(r.Context(), loginData)
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

}
