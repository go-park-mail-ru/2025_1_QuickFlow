package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"quickflow/shared/logger"
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
	FetchCreatorPosts(ctx context.Context, creatorId uuid.UUID, requesterId uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
	AddPost(ctx context.Context, post models.Post) (*models.Post, error)
	DeletePost(ctx context.Context, userId uuid.UUID, postId uuid.UUID) error
	UpdatePost(ctx context.Context, update models.PostUpdate, userId uuid.UUID) (*models.Post, error)
	LikePost(ctx context.Context, postId, userId uuid.UUID) error
	UnlikePost(ctx context.Context, postId, userId uuid.UUID) error
}

type FeedHandler struct {
	authUseCase      AuthUseCase
	postService      PostService
	profileUseCase   ProfileUseCase
	friendUseCase    FriendsUseCase
	communityService CommunityService
}

// NewFeedHandler creates new feed handler.
func NewFeedHandler(authUseCase AuthUseCase, postUseCase PostService,
	profileUseCase ProfileUseCase, friendUseCase FriendsUseCase,
	communityService CommunityService) *FeedHandler {
	return &FeedHandler{
		postService:      postUseCase,
		profileUseCase:   profileUseCase,
		friendUseCase:    friendUseCase,
		authUseCase:      authUseCase,
		communityService: communityService,
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

	posts, err := f.postService.FetchFeed(ctx, feedForm.Posts, ts, user.Id)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load feed: %s", err.Error()), appErr.HTTPStatus)
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
		info := forms.PublicUserInfoToOut(infosMap[authors[i]], rel)
		postsOut[i].Creator = &info
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

	posts, err := f.postService.FetchRecommendations(ctx, feedForm.Posts, ts, user.Id)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load recommendations: %s", err.Error()), appErr.HTTPStatus)
		return
	}

	var postsOut []forms.PostOut
	var authors []uuid.UUID
	var communities []uuid.UUID
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)
		postsOut = append(postsOut, postOut)
		if post.CreatorType == models.PostUser {
			authors = append(authors, post.CreatorId)
		} else {
			communities = append(communities, post.CreatorId)
		}
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

	infosCommunityMap := make(map[uuid.UUID]*models.Community)
	for _, communityId := range communities {
		infosCommunityMap[communityId], err = f.communityService.GetCommunityById(ctx, communityId)
		if err != nil {
			err := errors2.FromGRPCError(err)
			http2.WriteJSONError(w, fmt.Sprintf("Failed to load community info: %s", err.Error()), err.HTTPStatus)
			return
		}

		infosMap[infosCommunityMap[communityId].OwnerID], err = f.profileUseCase.GetPublicUserInfo(ctx, infosCommunityMap[communityId].OwnerID)
		if err != nil {
			err := errors2.FromGRPCError(err)
			http2.WriteJSONError(w, fmt.Sprintf("Failed to load user info: %s", err.Error()), err.HTTPStatus)
			return
		}
	}

	var numUser, numComm int
	for i := range postsOut {

		if postsOut[i].CreatorType == "user" {
			rel, err := f.friendUseCase.GetUserRelation(ctx, user.Id, authors[i-numComm])
			if err != nil {
				err := errors2.FromGRPCError(err)
				http2.WriteJSONError(w, fmt.Sprintf("Failed to get user relation: %s", err.Error()), err.HTTPStatus)
				return
			}

			info := forms.PublicUserInfoToOut(infosMap[authors[i-numComm]], rel)
			postsOut[i].Creator = &info
			numUser++
		} else {
			postsOut[i].Creator = forms.ToCommunityForm(*infosCommunityMap[communities[i-numUser]], infosMap[infosCommunityMap[communities[i-numUser]].OwnerID])
			numComm++
		}

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
	requester, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

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

	posts, err := f.postService.FetchCreatorPosts(ctx, user.Id, requester.Id, feedForm.Posts, ts)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load user posts: %s", err.Error()), appErr.HTTPStatus)
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
		info := forms.PublicUserInfoToOut(publicUserInfo, models.RelationSelf)
		postOut.Creator = &info
		postsOut = append(postsOut, postOut)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(postsOut)
	if err != nil {
		http2.WriteJSONError(w, "Failed to encode user posts", http.StatusInternalServerError)
	}
}

func (f *FeedHandler) FetchCommunityPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requester, ok := ctx.Value("user").(models.User)
	if !ok {
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	commName := mux.Vars(r)["name"]
	if len(commName) == 0 {
		http2.WriteJSONError(w, "Failed to get community ID from URL", http.StatusBadRequest)
		return
	}

	community, err := f.communityService.GetCommunityByName(ctx, commName)
	if err != nil {
		err := errors2.FromGRPCError(err)
		http2.WriteJSONError(w, fmt.Sprintf("Failed to get community: %s", err.Error()), err.HTTPStatus)
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

	posts, err := f.postService.FetchCreatorPosts(ctx, community.ID, requester.Id, feedForm.Posts, ts)
	appErr := errors2.FromGRPCError(err)
	if appErr != nil && appErr.HTTPStatus != http.StatusNotFound {
		http2.WriteJSONError(w, fmt.Sprintf("Failed to load community posts: %s", err.Error()), appErr.HTTPStatus)
		return
	}

	var postsOut []forms.PostOut
	for _, post := range posts {
		var postOut forms.PostOut
		postOut.FromPost(post)
		info, err := f.profileUseCase.GetPublicUserInfo(ctx, community.OwnerID)
		if err != nil {
			err := errors2.FromGRPCError(err)
			logger.Error(ctx, fmt.Sprintf("Failed to get user info: %s", err.Error()))
			http2.WriteJSONError(w, "Failed to get user info", err.HTTPStatus)
		}

		postOut.Creator = forms.ToCommunityForm(*community, info)
		postsOut = append(postsOut, postOut)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(postsOut)
	if err != nil {
		http2.WriteJSONError(w, "Failed to encode community posts", http.StatusInternalServerError)
	}
}
