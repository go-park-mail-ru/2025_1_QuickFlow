package models

import (
	"time"

	"github.com/google/uuid"
)

type Sex int

const (
	MALE = iota
	FEMALE
)

type User struct {
	Id       uuid.UUID
	Username string
	Password string
	Salt     string
	LastSeen time.Time
}
