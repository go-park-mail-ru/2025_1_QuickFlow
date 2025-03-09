package http

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "quickflow/config"
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

// NewFeedHandler creates new feed handler.
func NewFeedHandler(postUseCase PostUseCase, authUseCase AuthUseCase) *FeedHandler {
    return &FeedHandler{
        postUseCase: postUseCase,
        authUseCase: authUseCase,
    }
}

// AddPost adds post to the feed.
func (f *FeedHandler) AddPost(w http.ResponseWriter, r *http.Request) {
    // extracting user from context
    user, ok := r.Context().Value("user").(models.User)
    if !ok {
        WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
        return
    }

    // parsing JSON
    var postForm forms.PostForm
    err := json.NewDecoder(r.Body).Decode(&postForm)
    if err != nil {
        WriteJSONError(w, "Failed to parse JSON", http.StatusBadRequest)
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
        WriteJSONError(w, "Failed to add post", http.StatusInternalServerError)
        return
    }
    log.Printf("Post added: %v\n", post)
}

// GetFeed returns feed for user using JSON format
func (f *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
    // extracting user from context
    user, ok := r.Context().Value("user").(models.User)
    if !ok {
        WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
        return
    }

    // parsing JSON
    var feedForm forms.FeedForm
    err := json.NewDecoder(r.Body).Decode(&feedForm)
    if err != nil {
        WriteJSONError(w, "Failed to parse JSON", http.StatusBadRequest)
        return
    }

    ts, err := time.Parse(config.TimeStampLayout, feedForm.Ts)
    if err != nil {
        ts = time.Now()
    }

    posts, err := f.postUseCase.FetchFeed(r.Context(), user, feedForm.Posts, ts)
    if err != nil {
        WriteJSONError(w, "Failed to load feed", http.StatusInternalServerError)
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
        WriteJSONError(w, "Failed to encode feed", http.StatusInternalServerError)
        return
    }
}
