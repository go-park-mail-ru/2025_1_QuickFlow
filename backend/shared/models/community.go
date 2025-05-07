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

type BasicCommunityInfo struct {
	Name        string
	Description string
	AvatarUrl   string
	CoverUrl    string
}

type Community struct {
	ID       uuid.UUID
	NickName string

	BasicInfo   *BasicCommunityInfo
	Avatar      *File
	Cover       *File
	OwnerID     uuid.UUID
	CreatedAt   time.Time
	ContactInfo *ContactInfo
}

type CommunityMember struct {
	UserID      uuid.UUID
	CommunityID uuid.UUID
	Role        CommunityRole
	JoinedAt    time.Time
}
