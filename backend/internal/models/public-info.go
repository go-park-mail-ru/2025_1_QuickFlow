package models

import "github.com/google/uuid"

type PublicUserInfo struct {
	Id        uuid.UUID
	Username  string
	Firstname string
	Lastname  string
	AvatarURL string
}
