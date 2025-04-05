package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"quickflow/config"
	"quickflow/internal/delivery/forms"
	"quickflow/internal/models"
	http2 "quickflow/utils/http"
)

type PostUseCase interface {
	FetchFeed(ctx context.Context, user models.User, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) error
	DeletePost(ctx context.Context, user models.User, postId uuid.UUID) error
}

type FeedHandler struct {
	postUseCase PostUseCase
}

// NewFeedHandler creates new feed handler.
func NewFeedHandler(postUseCase PostUseCase) *FeedHandler {
	return &FeedHandler{
		postUseCase: postUseCase,
	}
}

// GetFeed возвращает ленту постов
// @Summary Получить ленту
// @Description Возвращает список постов, опубликованных до указанного времени
// @Tags Feed
// @Produce json
// @Param posts_count query int true "Количество постов"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.PostOut "Список постов"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/feed [get]
func (f *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	// extracting user from context
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	// parsing JSON
	var feedForm forms.FeedForm
	err := feedForm.GetParams(r.URL.Query())
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	ts, err := time.Parse(config.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postUseCase.FetchFeed(r.Context(), user, feedForm.Posts, ts)
	if err != nil {
		http2.WriteJSONError(w, "Failed to load feed", http.StatusInternalServerError)
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
		http2.WriteJSONError(w, "Failed to encode feed", http.StatusInternalServerError)
		return
	}
}
