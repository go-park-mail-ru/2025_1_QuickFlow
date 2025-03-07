package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
)

type PostUseCase interface {
	FetchFeed(ctx context.Context, user models.User, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) error
}

type FeedHandler struct {
	postUseCase PostUseCase
	authUseCase AuthUseCase
}

func NewFeedHandler(postUseCase PostUseCase, authUseCase AuthUseCase) *FeedHandler {
	return &FeedHandler{
		postUseCase: postUseCase,
		authUseCase: authUseCase,
	}
}

// AddPost adds post to the feed.
func (f *FeedHandler) AddPost(w http.ResponseWriter, r *http.Request) {
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

	user, err := f.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid})
	if err != nil {
		http.Error(w, "Failed to authorize user", http.StatusUnauthorized)
		return
	}
	var postForm forms.PostForm

	err = json.NewDecoder(r.Body).Decode(&postForm)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	post := models.Post{
		CreatorId: user.Id,
		Desc:      postForm.Desc,
		Pics:      postForm.Pics,
		CreatedAt: time.Now(),
	}

	err = f.postUseCase.AddPost(r.Context(), post)
	if err != nil {
		http.Error(w, "Failed to add post", http.StatusInternalServerError)
		return
	}
	log.Printf("Post added: %v\n", post)
}

// GetFeed returns feed for user using JSON format
func (f *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
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

	user, err := f.authUseCase.LookupUserSession(r.Context(), models.Session{SessionId: sessionUuid, ExpireDate: session.Expires})
	if err != nil {
		http.Error(w, "Failed to authorize user", http.StatusUnauthorized)
		return
	}

	var feedForm forms.FeedForm

	err = json.NewDecoder(r.Body).Decode(&feedForm)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// TODO: move to config
	const layout = "2006-01-02 15:04:05"

	// TODO: confirm layout with frontend
	ts, err := time.Parse(layout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postUseCase.FetchFeed(r.Context(), user, feedForm.Posts, ts)
	if err != nil {
		http.Error(w, "Failed to load feed", http.StatusInternalServerError)
		return
	}

	var postsOut []forms.PostOut
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)
		postsOut = append(postsOut, postOut)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(postsOut)
	if err != nil {
		http.Error(w, "Failed to encode feed", http.StatusInternalServerError)
		return
	}
}
