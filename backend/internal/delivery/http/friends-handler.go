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
	GetFriendsInfo(ctx context.Context, userID string) ([]models.FriendInfo, error)
}

type FriendsHandler struct {
	friendsUseCase FriendsUseCase
}

func NewFriendsHandler(friendsUseCase FriendsUseCase) *FriendsHandler {
	return &FriendsHandler{
		friendsUseCase: friendsUseCase,
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
	logger.Info(ctx, fmt.Sprintf("User %s requested friends", user.Login))

	friendsInfo, err := f.friendsUseCase.GetFriendsInfo(ctx, user.Id.String())
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get list of friends for user %s: %s", user.Login, err.Error()))
		http2.WriteJSONError(w, "Failed to get friends", http.StatusInternalServerError)
		return
	}

	var friendsInfoOut forms.FriendsInfoOut

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(friendsInfoOut.ToJson(friendsInfo))
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to encode friends info to json: %s", err))
		http2.WriteJSONError(w, "Unable to encode friends info to json", http.StatusInternalServerError)
		return
	}

	logger.Info(ctx, fmt.Sprintf("Successfully proccessed friends requset for user: %s", user.Login))

}
