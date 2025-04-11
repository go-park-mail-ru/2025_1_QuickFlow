package forms

import (
	"github.com/google/uuid"
	"quickflow/internal/models"
)

type FriendsInfoOut struct {
	ID         uuid.UUID `json:"id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"firstname"`
	LastName   string    `json:"lastname"`
	AvatarURL  string    `json:"avatar_url"`
	University string    `json:"university"`
}

func (f *FriendsInfoOut) toFriendsInfoOutForm(info models.FriendInfo) FriendsInfoOut {
	return FriendsInfoOut{
		ID:         info.Id,
		Username:   info.Username,
		FirstName:  info.Firstname,
		LastName:   info.Lastname,
		AvatarURL:  info.AvatarURL,
		University: info.University,
	}
}

func (f *FriendsInfoOut) ToJson(friendsInfo []models.FriendInfo) map[string]interface{} {
	res := make(map[string]interface{})

	var friendsInfoOut []FriendsInfoOut
	for _, friendInfo := range friendsInfo {
		friendsInfoOut = append(friendsInfoOut, f.toFriendsInfoOutForm(friendInfo))
	}

	res["body"] = friendsInfoOut

	return res

}
