package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/pkg/logger"
	http2 "quickflow/utils/http"
)

type FriendsUseCase interface {
	GetFriendsIds(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetFriendsInfo(ctx context.Context, friendIds []string) ([]models.FriendInfo, error)
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
// @Param posts_count query int true "Количество постов"
// @Param ts query string false "Временная метка"
// @Success 200 {array} forms.PostOut "Список постов"
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

	friendIDs, err := f.friendsUseCase.GetFriendsIds(ctx, user.Id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get list of friends for user %s: %s", user.Login, err.Error()))
		http2.WriteJSONError(w, "Failed to get friends", http.StatusInternalServerError)
	}

	friendsInfo, err := f.friendsUseCase.GetFriendsInfo(ctx, friendIDs)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to get list of friends for user %s: %s", user.Login, err.Error()))
		http2.WriteJSONError(w, "Failed to get friends", http.StatusInternalServerError)
	}

}
