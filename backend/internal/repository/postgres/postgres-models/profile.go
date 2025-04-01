package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"
	"quickflow/internal/models"
)

type ProfilePostgres struct {
	Id            pgtype.UUID
	Name          pgtype.Text
	Surname       pgtype.Text
	Sex           pgtype.Int4
	DateOfBirth   pgtype.Date
	Bio           pgtype.Text
	AvatarUrl     pgtype.Text
	BackgroundUrl pgtype.Text
}

// ConvertToProfile converts ProfilePostgres to models.Profile.
func (p *ProfilePostgres) ConvertToProfile() models.Profile {
	return models.Profile{
		UserId:        p.Id.Bytes,
		Name:          p.Name.String,
		Surname:       p.Surname.String,
		Sex:           models.Sex(p.Sex.Int32),
		DateOfBirth:   p.DateOfBirth.Time,
		Bio:           p.Bio.String,
		AvatarUrl:     p.AvatarUrl.String,
		BackgroundUrl: p.BackgroundUrl.String,
	}
}
