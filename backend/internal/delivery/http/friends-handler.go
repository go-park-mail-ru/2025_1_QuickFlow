package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"quickflow/internal/delivery/forms"

	"quickflow/internal/models"
	"quickflow/pkg/logger"
	http2 "quickflow/utils/http"
)

type FriendsUseCase interface {
	GetFriendsInfo(ctx context.Context, userID string, limit string, offset string) ([]models.FriendInfo, bool, int, error)
	SendFriendRequest(ctx context.Context, senderID string, receiverID string) error
	AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error
	Unfollow(ctx context.Context, userID string, friendID string) error
	DeleteFriend(ctx context.Context, user string, friend string) error
	IsExistsFriendRequest(ctx context.Context, senderID string, receiverID string) (bool, error)
	GetUserRelation(ctx context.Context, user1 uuid.UUID, user2 uuid.UUID) (models.UserRelation, error)
}

type FriendHandler struct {
	FriendsUseCase FriendsUseCase
	ConnService    IWebSocketConnectionManager
}

func NewFriendHandler(friendsUseCase FriendsUseCase, connService IWebSocketConnectionManager) *FriendHandler {
	return &FriendHandler{
		FriendsUseCase: friendsUseCase,
		ConnService:    connService,
	}
}

// GetFriends возвращает список друзей
// @Summary Получить друзей
// @Description Возвращает список друзей пользователя
// @Tags Friends
// @Produce json
// @Success 200 {array} forms.FriendsInfoOut "Список друзей"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/friends [get]
func (f *FriendHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	limit := r.URL.Query().Get("count")
	offset := r.URL.Query().Get("offset")
	userID := r.URL.Query().Get("user_id")

	targetUserID := userID
	if targetUserID == "" {
		targetUserID = user.Id.String()
	}

	logger.Info(ctx, fmt.Sprintf("User %s requested friends", targetUserID))

	friendsInfo, hasMore, friendsCount, err := f.FriendsUseCase.GetFriendsInfo(ctx, targetUserID, limit, offset)

	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get list of friends for user %s: %s", user.Username, err.Error()))
		http2.WriteJSONError(w, "Failed to get friends", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("Successfully get friends info for user %s", user.Username))

	var friendsOnline []bool
	for _, friend := range friendsInfo {
		_, isOnline := f.ConnService.IsConnected(friend.Id)
		friendsOnline = append(friendsOnline, isOnline)
	}

	var friendsInfoOut forms.FriendsInfoOut

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(friendsInfoOut.ToJson(friendsInfo, friendsOnline, hasMore, friendsCount))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to encode friends info to json: %s", err))
		http2.WriteJSONError(w, "Unable to encode friends info to json", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully proccessed friends requset for user: %s", user.Username))
}

// SendFriendRequest отправояет заявку в друзья
// @Summary Добавить в друзья
// @Description Отправляет запрос в друзья конкретному пользователю
// @Tags Friends
// @Produce json
// @Success 200 {array} forms.FriendsInfoOut "Список друзей"
// @Failure 400 {object} forms.ErrorForm "Некорректные данные"
// @Failure 409 {object} forms.ErrorForm "Отношение между пользователями (подписчик/друг) уже существует
// @Failure 500 {object} forms.ErrorForm "Ошибка сервера"
// @Router /api/friends [post]
func (f *FriendHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to decode request body: %s", err))
		http2.WriteJSONError(w, "Unable to decode request body", http.StatusBadRequest)
		return
	}

	logger.Info(ctx, "Successfully parsed request body")
	logger.Info(ctx, fmt.Sprintf("Trying to check relation between sender %s and receiver %s", user.Username, req.ReceiverID))

	isExists, err := f.FriendsUseCase.IsExistsFriendRequest(ctx, user.Id.String(), req.ReceiverID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to check relation between sender: %s and receiver: %s. Error: %s", user.Username, req.ReceiverID, err.Error()))
		http2.WriteJSONError(w, "Failed to check relation", http.StatusInternalServerError)
		return
	}

	if isExists {
		http2.WriteJSONError(w, "Relationship between sender and receiver already exists", http.StatusConflict)
		return
	}

	logger.Info(ctx, fmt.Sprintf("User %s trying to add friend %s ", user.Username, req.ReceiverID))

	if err = f.FriendsUseCase.SendFriendRequest(ctx, user.Id.String(), req.ReceiverID); err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to send friend request: %s", err))
		http2.WriteJSONError(w, "Failed to send friend request", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully processed friend request to user %s", req.ReceiverID))
}

func (f *FriendHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to decode request body: %s", err))
		http2.WriteJSONError(w, "Unable to decode request body", http.StatusBadRequest)
		return
	}

	logger.Info(ctx, "Successfully parsed request body")
	logger.Info(ctx, fmt.Sprintf("User %s trying to accept friend %s ", user.Username, req.ReceiverID))

	if err = f.FriendsUseCase.AcceptFriendRequest(ctx, user.Id.String(), req.ReceiverID); err != nil {
		http2.WriteJSONError(w, "Failed to accept friend request", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully accepted friend request from user %s", req.ReceiverID))
}

func (f *FriendHandler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequestDel
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to decode request body: %s", err))
		http2.WriteJSONError(w, "Unable to decode request body", http.StatusBadRequest)
		return
	}

	logger.Info(ctx, "Successfully parsed request body")
	logger.Info(ctx, fmt.Sprintf("User %s trying to delete friend %s ", user.Username, req.FriendID))

	if err = f.FriendsUseCase.DeleteFriend(ctx, user.Id.String(), req.FriendID); err != nil {
		http2.WriteJSONError(w, "Failed to accept friend request", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully deleted friend from user %s", req.FriendID))
}

func (f *FriendHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, "Trying to parse request body")

	var req forms.FriendRequestDel
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to decode request body: %s", err))
		http2.WriteJSONError(w, "Unable to decode request body", http.StatusBadRequest)
		return
	}

	logger.Info(ctx, "Successfully parsed request body")
	logger.Info(ctx, fmt.Sprintf("User: %s trying to unfollow userЖ %s ", user.Username, req.FriendID))

	if err = f.FriendsUseCase.Unfollow(ctx, user.Id.String(), req.FriendID); err != nil {
		http2.WriteJSONError(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully unfollowed friend from user %s", req.FriendID))
}
