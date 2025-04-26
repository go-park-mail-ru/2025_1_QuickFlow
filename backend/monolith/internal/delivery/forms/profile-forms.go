package forms

import (
	"errors"
	time2 "quickflow/monolith/config/time"
	models2 "quickflow/monolith/internal/models"
	"time"

	"github.com/google/uuid"
)

type ProfileInfo struct {
	Username      string     `json:"username,omitempty"`
	Name          string     `json:"firstname"`
	Surname     string      `json:"lastname"`
	Sex         models2.Sex `json:"sex"`
	DateOfBirth string      `json:"birth_date"`
	Bio           string     `json:"bio"`
	AvatarUrl     string     `json:"avatar_url,omitempty"`
	BackgroundUrl string     `json:"cover_url,omitempty"`
}

type ProfileForm struct {
	Id         string        `json:"id,omitempty"`
	Avatar     *models2.File `json:"-"`
	Background *models2.File `json:"-"`

	ProfileInfo         *ProfileInfo             `json:"profile"`
	ContactInfo         *ContactInfo             `json:"contact_info,omitempty"`
	SchoolEducation     *SchoolEducationForm     `json:"school,omitempty"`
	UniversityEducation *UniversityEducationForm `json:"university,omitempty"`
	LastSeen            string                   `json:"last_seen,omitempty"`
	IsOnline            *bool                    `json:"online,omitempty"`
	Relation            models2.UserRelation     `json:"relation,omitempty"`
	ChatId              *uuid.UUID               `json:"chat_id,omitempty"`
}

func (f *ProfileForm) FormToModel() (models2.Profile, error) {

	var contactInfo *models2.ContactInfo
	if f.ContactInfo != nil {
		contactInfo = &models2.ContactInfo{
			City:  f.ContactInfo.City,
			Email: f.ContactInfo.Email,
			Phone: f.ContactInfo.Phone,
		}
	}

	var basicInfo *models2.BasicInfo
	var err error
	if f.ProfileInfo != nil {
		basicInfo, err = ProfileInfoToModel(*f.ProfileInfo)
		if err != nil {
			return models2.Profile{}, err
		}
	}

	//id, err := uuid.Parse(f.Id)
	//if err != nil {
	//	return models.Profile{}, errors.New("invalid user id")
	//}

	return models2.Profile{
		//UserId:     id,
		Username:   f.ProfileInfo.Username,
		BasicInfo:  basicInfo,
		Avatar:     f.Avatar,
		Background: f.Background,

		SchoolEducation:     SchoolFormToModel(f.SchoolEducation),
		UniversityEducation: UniversityFormToModel(f.UniversityEducation),
		ContactInfo:         contactInfo,
	}, nil
}

func ModelToForm(profile models2.Profile, username string, isOnline bool, relation models2.UserRelation, uuid *uuid.UUID) ProfileForm {
	profileForm := ProfileForm{
		Id:                  profile.UserId.String(),
		ProfileInfo:         BasicInfoToForm(*profile.BasicInfo, username),
		SchoolEducation:     SchoolEducationToForm(profile.SchoolEducation),
		UniversityEducation: UniversityEducationToForm(profile.UniversityEducation),
		ContactInfo:         ContactInfoToForm(profile.ContactInfo),
		IsOnline:            &isOnline,
		Relation:            relation,
		ChatId:              uuid,
	}
	if !isOnline {
		profileForm.LastSeen = profile.LastSeen.Format(time2.TimeStampLayout)
	}
	return profileForm
}

type ContactInfo struct {
	City  string `json:"city,omitempty"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

func ContactInfoToForm(contactInfo *models2.ContactInfo) *ContactInfo {
	if contactInfo == nil {
		return nil
	}

	return &ContactInfo{
		City:  contactInfo.City,
		Email: contactInfo.Email,
		Phone: contactInfo.Phone,
	}
}

type SchoolEducationForm struct {
	SchoolCity string `json:"school_city,omitempty"`
	SchoolName string `json:"school_name,omitempty"`
}

type UniversityEducationForm struct {
	UniversityCity    string `json:"univ_city,omitempty"`
	UniversityName    string `json:"univ_name,omitempty"`
	UniversityFaculty string `json:"faculty,omitempty"`
	GraduationYear    int    `json:"grad_year,omitempty"`
}

type Activity struct {
	LastSeen string `json:"last_seen,omitempty"`
	IsOnline bool   `json:"online,omitempty"`
}

func SchoolEducationToForm(sch *models2.SchoolEducation) *SchoolEducationForm {
	if sch == nil {
		return nil
	}

	return &SchoolEducationForm{
		SchoolCity: sch.City,
		SchoolName: sch.School,
	}
}

func UniversityEducationToForm(uni *models2.UniversityEducation) *UniversityEducationForm {
	if uni == nil {
		return nil
	}

	return &UniversityEducationForm{
		UniversityCity:    uni.City,
		UniversityName:    uni.University,
		UniversityFaculty: uni.Faculty,
		GraduationYear:    uni.GraduationYear,
	}
}

func SchoolFormToModel(sch *SchoolEducationForm) *models2.SchoolEducation {
	if sch == nil {
		return nil
	}

	return &models2.SchoolEducation{
		City:   sch.SchoolCity,
		School: sch.SchoolName,
	}
}

func UniversityFormToModel(uniForm *UniversityEducationForm) *models2.UniversityEducation {
	if uniForm == nil {
		return nil
	}

	return &models2.UniversityEducation{
		City:           uniForm.UniversityCity,
		University:     uniForm.UniversityName,
		Faculty:        uniForm.UniversityFaculty,
		GraduationYear: uniForm.GraduationYear,
	}
}

func BasicInfoToForm(info models2.BasicInfo, username string) *ProfileInfo {
	return &ProfileInfo{
		Username:      username,
		Name:          info.Name,
		Surname:       info.Surname,
		Sex:           info.Sex,
		DateOfBirth:   info.DateOfBirth.Format(time2.DateLayout),
		Bio:           info.Bio,
		AvatarUrl:     info.AvatarUrl,
		BackgroundUrl: info.BackgroundUrl,
	}
}

func ProfileInfoToModel(info ProfileInfo) (*models2.BasicInfo, error) {
	date, err := time.Parse(time2.DateLayout, info.DateOfBirth)
	if err != nil {
		return nil, errors.New("incorrect date format")
	}
	return &models2.BasicInfo{
		Name:          info.Name,
		Surname:       info.Surname,
		Sex:           info.Sex,
		DateOfBirth:   date,
		Bio:           info.Bio,
		AvatarUrl:     info.AvatarUrl,
		BackgroundUrl: info.BackgroundUrl,
	}, nil
}
