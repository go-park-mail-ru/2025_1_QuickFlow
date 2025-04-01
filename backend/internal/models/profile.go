package models

import (
	"github.com/google/uuid"
	"time"
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

type Profile struct {
	UserId        uuid.UUID
	Name          string
	Surname       string
	Sex           Sex
	DateOfBirth   time.Time
	Bio           string
	AvatarUrl     string
	BackgroundUrl string
	Avatar        *File
	Background    *File

	ContactInfo         *ContactInfo
	SchoolEducation     *SchoolEducation
	UniversityEducation *UniversityEducation
}
