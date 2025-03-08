package postgres_models

import (
	"quickflow/config"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/internal/models"
)

type UserPostgres struct {
	Id          pgtype.UUID
	Login       pgtype.Text
	Name        pgtype.Text
	Surname     pgtype.Text
	Sex         pgtype.Int2
	DateOfBirth pgtype.Date
	Password    pgtype.Text
	Salt        pgtype.Text
}

// ConvertToUser converts UserPostgres to models.User.
func (u *UserPostgres) ConvertToUser() models.User {
	return models.User{
		Id:          u.Id.Bytes,
		Login:       u.Login.String,
		Name:        u.Name.String,
		Surname:     u.Surname.String,
		DateOfBirth: u.DateOfBirth.Time.String(),
		Sex:         models.Sex(u.Sex.Int16),
		Password:    u.Password.String,
		Salt:        u.Salt.String,
	}
}

func convertStringToPostgresText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// ConvertUserToPostgres converts models.User to UserPostgres.
func ConvertUserToPostgres(u models.User) UserPostgres {
	var dob pgtype.Date
	t, err := time.Parse(config.DateLayout, u.DateOfBirth)
	if err == nil {
		dob = pgtype.Date{Time: t, Valid: true}
	} else {
		dob = pgtype.Date{Valid: false} // NULL
	}

	return UserPostgres{
		Id:          pgtype.UUID{Bytes: u.Id, Valid: true},
		Login:       pgtype.Text{String: u.Login, Valid: true},
		Name:        convertStringToPostgresText(u.Name),
		Surname:     convertStringToPostgresText(u.Surname),
		Sex:         pgtype.Int2{Int16: int16(u.Sex), Valid: true},
		DateOfBirth: dob,
		Password:    pgtype.Text{String: u.Password, Valid: true},
		Salt:        pgtype.Text{String: u.Salt, Valid: true},
	}
}
