package forms

import (
	"errors"
	"quickflow/config"
	"quickflow/internal/models"
	"time"
)

type ProfileForm struct {
	Name          string       `json:"firstname"`
	Surname       string       `json:"lastname"`
	Sex           models.Sex   `json:"sex"`
	DateOfBirth   string       `json:"birth_date"`
	Bio           string       `json:"bio"`
	AvatarUrl     string       `json:"avatar_url,omitempty"`
	BackgroundUrl string       `json:"background_url,omitempty"`
	Avatar        *models.File `json:"-"`
	Background    *models.File `json:"-"`
}

func (f *ProfileForm) FormToModel() (models.Profile, error) {
	date, err := time.Parse(config.DateLayout, f.DateOfBirth)
	if err != nil {
		return models.Profile{}, errors.New("incorrect date format")
	}

	return models.Profile{
		Name:          f.Name,
		Surname:       f.Surname,
		Sex:           f.Sex,
		DateOfBirth:   date,
		Bio:           f.Bio,
		AvatarUrl:     f.AvatarUrl,
		Avatar:        f.Avatar,
		Background:    f.Background,
		BackgroundUrl: f.BackgroundUrl,
	}, nil
}

func ModelToForm(profile models.Profile) ProfileForm {
	return ProfileForm{
		Name:          profile.Name,
		Surname:       profile.Surname,
		Sex:           profile.Sex,
		DateOfBirth:   profile.DateOfBirth.Format(config.DateLayout),
		Bio:           profile.Bio,
		AvatarUrl:     profile.AvatarUrl,
		BackgroundUrl: profile.BackgroundUrl,
	}
}
