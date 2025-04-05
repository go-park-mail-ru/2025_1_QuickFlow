package models

import (
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
	UserId uuid.UUID

	BasicInfo           *BasicInfo
	ContactInfo         *ContactInfo
	SchoolEducation     *SchoolEducation
	UniversityEducation *UniversityEducation

	Avatar     *File
	Background *File
}
