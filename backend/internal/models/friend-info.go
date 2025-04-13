package models

import "github.com/google/uuid"

type FriendInfo struct {
	Id         uuid.UUID
	Username   string
	Firstname  string
	Lastname   string
	AvatarURL  string
	University string
}

type UserRelation string

const (
	RelationFriend     UserRelation = "friend"
	RelationFollowing  UserRelation = "following"
	RelationFollowedBy UserRelation = "followed_by"
)
