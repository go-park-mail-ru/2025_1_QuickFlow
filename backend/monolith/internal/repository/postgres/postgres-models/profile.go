package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"
	models2 "quickflow/monolith/internal/models"
)

type SchoolEducation struct {
	City   pgtype.Text
	School pgtype.Text
}

type UniversityEducation struct {
	City           pgtype.Text
	University     pgtype.Text
	Faculty        pgtype.Text
	GraduationYear pgtype.Int4
}

type ContactInfoPostgres struct {
	City  pgtype.Text
	Email pgtype.Text
	Phone pgtype.Text
}
type ProfilePostgres struct {
	Id            pgtype.UUID
	Name          pgtype.Text
	Surname       pgtype.Text
	Sex           pgtype.Int4
	DateOfBirth   pgtype.Date
	Bio           pgtype.Text
	AvatarUrl     pgtype.Text
	BackgroundUrl pgtype.Text
	LastSeen      pgtype.Timestamptz

	ContactInfo         *ContactInfoPostgres
	SchoolEducation     *SchoolEducation
	UniversityEducation *UniversityEducation
}

// ConvertToProfile converts ProfilePostgres to models.Profile.
func (p *ProfilePostgres) ConvertToProfile() models2.Profile {
	var profile models2.Profile

	if p.ContactInfo != nil {
		profile.ContactInfo = &models2.ContactInfo{
			City:  p.ContactInfo.City.String,
			Email: p.ContactInfo.Email.String,
			Phone: p.ContactInfo.Phone.String,
		}
	}

	if p.SchoolEducation != nil {
		profile.SchoolEducation = &models2.SchoolEducation{
			City:   p.SchoolEducation.City.String,
			School: p.SchoolEducation.School.String,
		}
	}

	if p.UniversityEducation != nil {
		profile.UniversityEducation = &models2.UniversityEducation{
			City:           p.UniversityEducation.City.String,
			University:     p.UniversityEducation.University.String,
			Faculty:        p.UniversityEducation.Faculty.String,
			GraduationYear: int(p.UniversityEducation.GraduationYear.Int32),
		}
	}

	return models2.Profile{
		UserId: p.Id.Bytes,
		BasicInfo: &models2.BasicInfo{
			Name:          p.Name.String,
			Surname:       p.Surname.String,
			Sex:           models2.Sex(p.Sex.Int32),
			DateOfBirth:   p.DateOfBirth.Time,
			Bio:           p.Bio.String,
			AvatarUrl:     p.AvatarUrl.String,
			BackgroundUrl: p.BackgroundUrl.String,
		},

		ContactInfo:         profile.ContactInfo,
		SchoolEducation:     profile.SchoolEducation,
		UniversityEducation: profile.UniversityEducation,
		LastSeen:            p.LastSeen.Time,
	}
}

type PublicUserInfoPostgres struct {
	Id        pgtype.UUID
	Username  pgtype.Text
	Firstname pgtype.Text
	Lastname  pgtype.Text
	AvatarURL pgtype.Text
	LastSeen  pgtype.Timestamptz
}

func (p *PublicUserInfoPostgres) ConvertToPublicUserInfo() models2.PublicUserInfo {
	return models2.PublicUserInfo{
		Id:        p.Id.Bytes,
		Username:  p.Username.String,
		Firstname: p.Firstname.String,
		Lastname:  p.Lastname.String,
		AvatarURL: p.AvatarURL.String,
		LastSeen:  p.LastSeen.Time,
	}
}

type FriendInfoPostgres struct {
	Id         pgtype.UUID
	Username   pgtype.Text
	Firstname  pgtype.Text
	Lastname   pgtype.Text
	AvatarURL  pgtype.Text
	University pgtype.Text
}

func (f *FriendInfoPostgres) ConvertToFriendInfo() models2.FriendInfo {
	return models2.FriendInfo{
		Id:         f.Id.Bytes,
		Username:   f.Username.String,
		Firstname:  f.Firstname.String,
		Lastname:   f.Lastname.String,
		AvatarURL:  f.AvatarURL.String,
		University: f.University.String,
	}
}
