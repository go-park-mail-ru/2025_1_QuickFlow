package http

import (
    "context"
    "encoding/json"
    "net/http"
    http2 "quickflow/utils/http"
    "quickflow/utils/validation"
    "time"

    "github.com/google/uuid"

    "quickflow/internal/delivery/forms"
    "quickflow/internal/models"
)

type AuthUseCase interface {
    CreateUser(ctx context.Context, user models.User) (uuid.UUID, models.Session, error)
    GetUser(ctx context.Context, authData models.LoginData) (models.Session, error)
    LookupUserSession(ctx context.Context, session models.Session) (models.User, error)
    DeleteUserSession(ctx context.Context, session string) error
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
        http2.WriteJSONError(w, "Bad request", http.StatusBadRequest)
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
    if err := validation.Validate(user.Login, user.Password, user.Name, user.Surname); err != nil {
        http2.WriteJSONError(w, err.Error(), http.StatusBadRequest)
        return
    }

    // process data
    id, session, err := a.authUseCase.CreateUser(r.Context(), user)
    if err != nil {
        http2.WriteJSONError(w, err.Error(), http.StatusConflict)
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
        http2.WriteJSONError(w, "Bad request", http.StatusBadRequest)
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
        http2.WriteJSONError(w, "Unauthorized", http.StatusUnauthorized)
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

func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    cookieUUID, err := uuid.Parse(cookie.Value)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if _, err = a.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: cookieUUID}); err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if err = a.authUseCase.DeleteUserSession(r.Context(), cookie.Value); err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    cookie.Expires = time.Now().AddDate(0, 0, -1)
    http.SetCookie(w, cookie)
}
