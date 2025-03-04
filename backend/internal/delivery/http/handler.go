package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/utils"
)

type useCase interface {
	FetchFeed(ctx context.Context, user models.User, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) error
	LookupUserSession(ctx context.Context, session models.Session) (models.User, error)

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
	_, err := w.Write([]byte("Hello, world!\n"))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// FOR TESTING ONLY
type PostForm struct {
	Desc string   `json:"desc"`
	Pic  []string `json:"pic"`
}

// AddPost adds post to the feed.
func (h *Handler) AddPost(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	session, err := r.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		http.Error(w, "Authorization needed", http.StatusUnauthorized)
		return
	}

	sessionUuid, err := uuid.Parse(session.Value)
	if err != nil {
		http.Error(w, "Failed to parse session", http.StatusBadRequest)
		return
	}

	user, err := h.useCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid, ExpireDate: session.Expires})
	if err != nil {
		http.Error(w, "Failed to authorize user", http.StatusUnauthorized)
		return
	}
	var postForm PostForm

	err = json.NewDecoder(r.Body).Decode(&postForm)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	post := models.Post{
		CreatorId: user.Id,
		Desc:      postForm.Desc,
		Pics:      postForm.Pic,
		CreatedAt: time.Now(),
	}

	err = h.useCase.AddPost(r.Context(), post)
	if err != nil {
		http.Error(w, "Failed to add post", http.StatusInternalServerError)
		return
	}

	// TODO: redirect to post page
	log.Printf("Post added: %v\n", post)
}

// GetFeed returns feed for user using JSON format
func (h *Handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		http.Error(w, "Authorization needed", http.StatusUnauthorized)
		return
	}

	sessionUuid, err := uuid.Parse(session.Value)
	if err != nil {
		http.Error(w, "Failed to parse session", http.StatusBadRequest)
		return
	}

	user, err := h.useCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid, ExpireDate: session.Expires})
	if err != nil {
		http.Error(w, "Failed to authorize user", http.StatusUnauthorized)
		return
	}

	numPosts, err := strconv.ParseInt(r.URL.Query().Get("posts"), 10, 32)
	if err != nil {
		http.Error(w, "Failed to parse number of posts", http.StatusBadRequest)
		return
	}

	// TODO: confirm layout with frontend
	ts, err := time.Parse(time.Layout, r.URL.Query().Get("ts"))

	if err != nil {
		ts = time.Now()
	}

	posts, err := h.useCase.FetchFeed(r.Context(), user, int(numPosts), ts)
	if err != nil {
		http.Error(w, "Failed to load feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		http.Error(w, "Failed to encode feed", http.StatusInternalServerError)
		return
	}
}

// SignUp creates new user.
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

// Login logs in user.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var authData models.AuthForm

	if err := json.NewDecoder(r.Body).Decode(&authData); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// process data
	session, err := h.useCase.GetUser(authData)
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
