package forms

import (
	"errors"
	"quickflow/config"
	"quickflow/internal/models"
	"time"
)

type ProfileForm struct {
	Username      string       `json:"username,omitempty"`
	Name          string       `json:"firstname"`
	Surname       string       `json:"lastname"`
	Sex           models.Sex   `json:"sex"`
	DateOfBirth   string       `json:"birth_date"`
	Bio           string       `json:"bio"`
	AvatarUrl     string       `json:"avatar_url,omitempty"`
	BackgroundUrl string       `json:"cover_url,omitempty"`
	Avatar        *models.File `json:"-"`
	Background    *models.File `json:"-"`

	ContactInfo         *ContactInfo             `json:"contact_info,omitempty"`
	SchoolEducation     *SchoolEducationForm     `json:"school_education,omitempty"`
	UniversityEducation *UniversityEducationForm `json:"university_education,omitempty"`
}

func (f *ProfileForm) FormToModel() (models.Profile, error) {
	date, err := time.Parse(config.DateLayout, f.DateOfBirth)
	if err != nil {
		return models.Profile{}, errors.New("incorrect date format")
	}
	var contactInfo *models.ContactInfo
	if f.ContactInfo != nil {
		contactInfo = &models.ContactInfo{
			City:  f.ContactInfo.City,
			Email: f.ContactInfo.Email,
			Phone: f.ContactInfo.Phone,
		}
	}

	return models.Profile{
		Name:                f.Name,
		Surname:             f.Surname,
		Sex:                 f.Sex,
		DateOfBirth:         date,
		Bio:                 f.Bio,
		AvatarUrl:           f.AvatarUrl,
		Avatar:              f.Avatar,
		Background:          f.Background,
		BackgroundUrl:       f.BackgroundUrl,
		SchoolEducation:     SchoolFormToModel(f.SchoolEducation),
		UniversityEducation: UniversityFormToModel(f.UniversityEducation),
		ContactInfo:         contactInfo,
	}, nil
}

func ModelToForm(profile models.Profile, username string) ProfileForm {
	return ProfileForm{
		Username:      username,
		Name:          profile.Name,
		Surname:       profile.Surname,
		Sex:           profile.Sex,
		DateOfBirth:   profile.DateOfBirth.Format(config.DateLayout),
		Bio:           profile.Bio,
		AvatarUrl:     profile.AvatarUrl,
		BackgroundUrl: profile.BackgroundUrl,

		SchoolEducation:     SchoolEducationToForm(profile.SchoolEducation),
		UniversityEducation: UniversityEducationToForm(profile.UniversityEducation),
		ContactInfo:         ContactInfoToForm(profile.ContactInfo),
	}
}

type ContactInfo struct {
	City  string `json:"city,omitempty"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

func ContactInfoToForm(contactInfo *models.ContactInfo) *ContactInfo {
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

func SchoolEducationToForm(sch *models.SchoolEducation) *SchoolEducationForm {
	if sch == nil {
		return nil
	}

	return &SchoolEducationForm{
		SchoolCity: sch.City,
		SchoolName: sch.School,
	}
}

func UniversityEducationToForm(uni *models.UniversityEducation) *UniversityEducationForm {
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

func SchoolFormToModel(sch *SchoolEducationForm) *models.SchoolEducation {
	if sch == nil {
		return nil
	}

	return &models.SchoolEducation{
		City:   sch.SchoolCity,
		School: sch.SchoolName,
	}
}

func UniversityFormToModel(uniForm *UniversityEducationForm) *models.UniversityEducation {
	if uniForm == nil {
		return nil
	}

	return &models.UniversityEducation{
		City:           uniForm.UniversityCity,
		University:     uniForm.UniversityName,
		Faculty:        uniForm.UniversityFaculty,
		GraduationYear: uniForm.GraduationYear,
	}
}
