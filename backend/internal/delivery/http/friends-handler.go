package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"quickflow/internal/delivery/forms"

	"quickflow/internal/models"
	"quickflow/pkg/logger"
	http2 "quickflow/utils/http"
)

type FriendsUseCase interface {
	GetFriendsInfo(ctx context.Context, userID string, limit string, offset string) ([]models.FriendInfo, bool, error)
	SendFriendRequest(ctx context.Context, senderID string, receiverID string) error
	AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error
	IsExistsFriendRequest(ctx context.Context, senderID string, receiverID string) (bool, error)
}

type FriendsHandler struct {
	friendsUseCase FriendsUseCase
	connService    IWebSocketManager
}

func NewFriendsHandler(friendsUseCase FriendsUseCase, connService IWebSocketManager) *FriendsHandler {
	return &FriendsHandler{
		friendsUseCase: friendsUseCase,
		connService:    connService,
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
func (f *FriendsHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("User %s requested friends", user.Username))

	limit, offset := r.URL.Query().Get("count"), r.URL.Query().Get("offset")

	friendsInfo, hasMore, err := f.friendsUseCase.GetFriendsInfo(ctx, user.Id.String(), limit, offset)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get list of friends for user %s: %s", user.Username, err.Error()))
		http2.WriteJSONError(w, "Failed to get friends", http.StatusInternalServerError)
		return
	}
	logger.Info(ctx, fmt.Sprintf("Successfully get friends info for user %s", user.Username))

	var friendsOnline []bool
	for _, friend := range friendsInfo {
		_, isOnline := f.connService.IsConnected(friend.Id)
		friendsOnline = append(friendsOnline, isOnline)
	}

	var friendsInfoOut forms.FriendsInfoOut

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(friendsInfoOut.ToJson(friendsInfo, friendsOnline, hasMore))
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
func (f *FriendsHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	var req forms.FriendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to decode request body: %s", err))
		http2.WriteJSONError(w, "Unable to decode request body", http.StatusBadRequest)
		return
	}

	isExists, err := f.friendsUseCase.IsExistsFriendRequest(ctx, user.Id.String(), req.ReceiverID)
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

	if err = f.friendsUseCase.SendFriendRequest(ctx, user.Id.String(), req.ReceiverID); err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to send friend request: %s", err))
		http2.WriteJSONError(w, "Failed to send friend request", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully processed friend request to user %s", req.ReceiverID))
}

func (f *FriendsHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	ctx := http2.SetRequestId(r.Context())

	user, ok := ctx.Value("user").(models.User)
	if !ok {
		logger.Error(ctx, "Failed to get user from context while fetching friends")
		http2.WriteJSONError(w, "Failed to get user from context", http.StatusInternalServerError)
		return
	}

	var req forms.FriendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to decode request body: %s", err))
		http2.WriteJSONError(w, "Unable to decode request body", http.StatusBadRequest)
		return
	}

	if err = f.friendsUseCase.AcceptFriendRequest(ctx, user.Id.String(), req.ReceiverID); err != nil {
		http2.WriteJSONError(w, "Failed to accept friend request", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully accepted friend request from user %s", req.ReceiverID))

}
