package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type UserPostgres struct {
	Id       pgtype.UUID
	Username pgtype.Text
	Password pgtype.Text
	Salt     pgtype.Text
}

// ConvertToUser converts UserPostgres to models.User.
func (u *UserPostgres) ConvertToUser() models.User {
	return models.User{
		Id:       u.Id.Bytes,
		Username: u.Username.String,
		Password: u.Password.String,
		Salt:     u.Salt.String,
	}
}

func convertStringToPostgresText(s string) pgtype.Text {
	if len(s) == 0 {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// ConvertUserToPostgres converts models.User to UserPostgres.
func ConvertUserToPostgres(u models.User) UserPostgres {

	return UserPostgres{
		Id:       pgtype.UUID{Bytes: u.Id, Valid: true},
		Username: pgtype.Text{String: u.Username, Valid: true},
		Password: pgtype.Text{String: u.Password, Valid: true},
		Salt:     pgtype.Text{String: u.Salt, Valid: true},
	}
}
