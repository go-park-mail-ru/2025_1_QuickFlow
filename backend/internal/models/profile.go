package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SchoolEducation struct {
	City   string
	School string
}

type UniversityEducation struct {
	City           string
	University     string
	Faculty        string
	GraduationYear int
}

type ContactInfo struct {
	City  string
	Email string
	Phone string
}

type BasicInfo struct {
	Name          string
	Surname       string
	Sex           Sex
	DateOfBirth   time.Time
	Bio           string
	AvatarUrl     string
	BackgroundUrl string
}

type Profile struct {
	UserId   uuid.UUID
	Username string

	BasicInfo           *BasicInfo
	ContactInfo         *ContactInfo
	SchoolEducation     *SchoolEducation
	UniversityEducation *UniversityEducation

	Avatar     *File
	Background *File
	LastSeen   time.Time
}

func (p Profile) String() string {
	return fmt.Sprintf(
		"Profile{UserId: %s, BasicInfo: %s, ContactInfo: %s, SchoolEducation: %s, UniversityEducation: %s, Avatar: %s, Background: %s}",
		p.UserId,
		formatIfNotNil(p.BasicInfo),
		formatIfNotNil(p.ContactInfo),
		formatIfNotNil(p.SchoolEducation),
		formatIfNotNil(p.UniversityEducation),
		formatIfNotNil(p.Avatar),
		formatIfNotNil(p.Background),
	)
}

func formatIfNotNil(v interface{}) string {
	if v == nil {
		return "nil"
	}
	return fmt.Sprintf("%+v", v)
}
