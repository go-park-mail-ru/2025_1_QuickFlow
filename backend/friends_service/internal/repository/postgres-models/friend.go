package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type FriendInfoPostgres struct {
	Id         pgtype.UUID
	Username   pgtype.Text
	Firstname  pgtype.Text
	Lastname   pgtype.Text
	AvatarURL  pgtype.Text
	University pgtype.Text
}

func (f *FriendInfoPostgres) ConvertToFriendInfo() models.FriendInfo {
	return models.FriendInfo{
		Id:         f.Id.Bytes,
		Username:   f.Username.String,
		Firstname:  f.Firstname.String,
		Lastname:   f.Lastname.String,
		AvatarURL:  f.AvatarURL.String,
		University: f.University.String,
	}
}
