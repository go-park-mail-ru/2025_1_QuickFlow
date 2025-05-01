package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	time2 "quickflow/config/time"
	"quickflow/gateway/internal/delivery/forms"
	errors2 "quickflow/gateway/internal/errors"
	http2 "quickflow/gateway/utils/http"
	"quickflow/shared/models"
)

type PostService interface {
	FetchFeed(ctx context.Context, numPosts int, timestamp time.Time, userId uuid.UUID) ([]models.Post, error)
	FetchRecommendations(ctx context.Context, numPosts int, timestamp time.Time, userId uuid.UUID) ([]models.Post, error)
	FetchUserPosts(ctx context.Context, userId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, userId uuid.UUID, postId uuid.UUID) error
	UpdatePost(ctx context.Context, update models.PostUpdate, userId uuid.UUID) (*models.Post, error)
}

type FeedHandler struct {
	authUseCase    AuthUseCase
	postUseCase    PostService
	profileUseCase ProfileUseCase
	friendUseCase  FriendsUseCase
}

// NewFeedHandler creates new feed handler.
func NewFeedHandler(authUseCase AuthUseCase, postUseCase PostService, profileUseCase ProfileUseCase, friendUseCase FriendsUseCase) *FeedHandler {
	return &FeedHandler{
		postUseCase:    postUseCase,
		profileUseCase: profileUseCase,
		friendUseCase:  friendUseCase,
		authUseCase:    authUseCase,
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
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	var feedForm forms.FeedForm
	err := feedForm.GetParams(r.URL.Query())
	if err != nil {
		http2.WriteJSONError(w, fmt.Sprintf("Failed to parse query params: %s", err.Error()), http.StatusBadRequest)
		return
	}

	ts, err := time.Parse(time2.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postUseCase.FetchFeed(ctx, feedForm.Posts, ts, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load feed: %s", err.Error()), err.HTTPStatus)
		return
	}

	var postsOut []forms.PostOut
	var authors []uuid.UUID
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)
		postsOut = append(postsOut, postOut)
		authors = append(authors, post.CreatorId)
	}

	publicAuthorsInfo, err := f.profileUseCase.GetPublicUsersInfo(ctx, authors)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load authors info: %s", err.Error()), err.HTTPStatus)
		return
	}

	infosMap := make(map[uuid.UUID]models.PublicUserInfo)
	for _, info := range publicAuthorsInfo {
		infosMap[info.Id] = info
	}

	for i := range postsOut {
		rel, err := f.friendUseCase.GetUserRelation(ctx, user.Id, authors[i])
		if err != nil {
			err := errors2.FromGRPCError(err)
			http2.WriteJSONError(w, fmt.Sprintf("Failed to get user relation: %s", err.Error()), err.HTTPStatus)
			return
		}
		postsOut[i].Creator = forms.PublicUserInfoToOut(infosMap[authors[i]], rel)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(postsOut)
	if err != nil {
		http2.WriteJSONError(w, "Failed to encode feed", http.StatusInternalServerError)
	}
}

// GetRecommendations возвращает рекомендации
// @Summary Получить рекомендации
// @Description Возвращает список постов, опубликованных до указанного времени
// @Tags Feed
// @Produce json
// @Param posts_count query int true "Количество постов"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.PostOut "Список постов"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/recommendations [get]
// @Security Session
func (f *FeedHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	var feedForm forms.FeedForm
	err := feedForm.GetParams(r.URL.Query())
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	ts, err := time.Parse(time2.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postUseCase.FetchRecommendations(ctx, feedForm.Posts, ts, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load recommendations: %s", err.Error()), err.HTTPStatus)
		return
	}

	var postsOut []forms.PostOut
	var authors []uuid.UUID
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)
		postsOut = append(postsOut, postOut)
		authors = append(authors, post.CreatorId)
	}

	publicAuthorsInfo, err := f.profileUseCase.GetPublicUsersInfo(ctx, authors)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load authors info: %s", err.Error()), err.HTTPStatus)
		return
	}

	infosMap := make(map[uuid.UUID]models.PublicUserInfo)
	for _, info := range publicAuthorsInfo {
		infosMap[info.Id] = info
	}

	for i := range postsOut {
		rel, err := f.friendUseCase.GetUserRelation(ctx, user.Id, authors[i])
		if err != nil {
			err := errors2.FromGRPCError(err)
			http2.WriteJSONError(w, fmt.Sprintf("Failed to get user relation: %s", err.Error()), err.HTTPStatus)
			return
		}
		postsOut[i].Creator = forms.PublicUserInfoToOut(infosMap[authors[i]], rel)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(postsOut)
	if err != nil {
		http2.WriteJSONError(w, "Failed to encode recommendations", http.StatusInternalServerError)
	}
}

// FetchUserPosts возвращает посты пользователя
// @Summary Получить посты пользователя
// @Description Возвращает список постов пользователя, опубликованных до указанного времени
// @Tags Feed
// @Produce json
// @Param posts_count query int true "Количество постов"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.PostOut "Список постов"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/profiles/{username}/posts [get]
// @Security Session
func (f *FeedHandler) FetchUserPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := mux.Vars(r)["username"]
	if username == "" {
		http2.WriteJSONError(w, "Failed to get username from URL", http.StatusBadRequest)
		return
	}

	user, err := f.authUseCase.GetUserByUsername(ctx, username)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to get user: %s", err.Error()), err.HTTPStatus)
		return
	}

	var feedForm forms.FeedForm
	err = feedForm.GetParams(r.URL.Query())
	if err != nil {
		http2.WriteJSONError(w, "Failed to parse query params", http.StatusBadRequest)
		return
	}

	ts, err := time.Parse(time2.TimeStampLayout, feedForm.Ts)
	if err != nil {
		ts = time.Now()
	}

	posts, err := f.postUseCase.FetchUserPosts(ctx, user.Id, feedForm.Posts, ts)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load user posts: %s", err.Error()), err.HTTPStatus)
		return
	}

	publicUserInfo, err := f.profileUseCase.GetPublicUserInfo(ctx, user.Id)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load user info: %s", err.Error()), err.HTTPStatus)
		return
	}

	var postsOut []forms.PostOut
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)
		postOut.Creator = forms.PublicUserInfoToOut(publicUserInfo, models.RelationSelf)
		postsOut = append(postsOut, postOut)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(postsOut)
	if err != nil {
		http2.WriteJSONError(w, "Failed to encode user posts", http.StatusInternalServerError)
	}
}
