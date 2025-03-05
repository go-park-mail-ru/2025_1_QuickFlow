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
)

type PostUseCase interface {
	FetchFeed(ctx context.Context, user models.User, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) error
}

type PostHandler struct {
	postUseCase PostUseCase
	authUseCase AuthUseCase
}

func NewPostHandler(postUseCase PostUseCase, authUseCase AuthUseCase) *PostHandler {
	return &PostHandler{
		postUseCase: postUseCase,
		authUseCase: authUseCase,
	}
}

// TODO: FOR TESTING ONLY. TRANSPORT STRUCTURES NEEDED.
type PostForm struct {
	Desc string   `json:"desc"`
	Pic  []string `json:"pic"`
}

// AddPost adds post to the feed.
func (p *PostHandler) AddPost(w http.ResponseWriter, r *http.Request) {
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

	user, err := p.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid, ExpireDate: session.Expires})
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

	err = p.postUseCase.AddPost(r.Context(), post)
	if err != nil {
		http.Error(w, "Failed to add post", http.StatusInternalServerError)
		return
	}

	// TODO: redirect to post page
	log.Printf("Post added: %v\n", post)
}

// GetFeed returns feed for user using JSON format
func (h *PostHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid, ExpireDate: session.Expires})
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

	posts, err := h.postUseCase.FetchFeed(r.Context(), user, int(numPosts), ts)
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
