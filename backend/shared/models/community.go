package models

import (
	"time"

	"github.com/google/uuid"
)

type CommunityRole string

const (
	CommunityRoleMember CommunityRole = "member"
	CommunityRoleAdmin  CommunityRole = "admin"
	CommunityRoleOwner  CommunityRole = "owner"
)

type Community struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	Avatar      *File
	AvatarUrl   string
	OwnerID     uuid.UUID
}

type CommunityMember struct {
	UserID      uuid.UUID
	CommunityID uuid.UUID
	Role        CommunityRole
	JoinedAt    time.Time
}
