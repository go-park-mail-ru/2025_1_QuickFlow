package forms

import (
	"github.com/google/uuid"

	"quickflow/internal/models"
)

type FriendRequest struct {
	ReceiverID string `json:"receiver_id"`
}

type FriendRequestDel struct {
	FriendID string `json:"friend_id"`
}

type FriendsInfoOut struct {
	ID         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"firstname"`
	LastName   string    `json:"lastname"`
	AvatarURL  string    `json:"avatar_url"`
	University string    `json:"university"`
	IsOnline   bool      `json:"is_online"`
}

func (f *FriendsInfoOut) toFriendsInfoOutForm(info models.FriendInfo, isOnline bool) FriendsInfoOut {
	return FriendsInfoOut{
		ID:         info.Id,
		Username:   info.Username,
		FirstName:  info.Firstname,
		LastName:   info.Lastname,
		AvatarURL:  info.AvatarURL,
		University: info.University,
		IsOnline:   isOnline,
	}
}

func (f *FriendsInfoOut) ToJson(friendsInfo []models.FriendInfo, friendsOnline []bool, hasMore bool, friendsCount int) map[string]map[string]interface{} {
	res := make(map[string]map[string]interface{})
	res["body"] = make(map[string]interface{})

	var friendsInfoOut []FriendsInfoOut
	for i, friendInfo := range friendsInfo {
		friendsInfoOut = append(friendsInfoOut, f.toFriendsInfoOutForm(friendInfo, friendsOnline[i]))
	}

	res["body"]["friends"] = friendsInfoOut
	res["body"]["has_more"] = hasMore
	res["body"]["total_count"] = friendsCount

	return res

}
