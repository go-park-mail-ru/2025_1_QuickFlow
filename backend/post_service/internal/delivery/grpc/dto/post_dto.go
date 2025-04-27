package dto

import (
	"time"

	"github.com/google/uuid"
)

type PostDTO struct {
	Id           uuid.UUID
	CreatorId    uuid.UUID
	Description  string
	ImagesURL    []string
	Images       []*FileDTO
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LikeCount    int
	RepostCount  int
	CommentCount int
	IsRepost     bool
}

type AddPostRequestDTO struct {
	Post *PostDTO
}

type AddPostResponseDTO struct {
	Post *PostDTO
}

type DeletePostRequestDTO struct {
	PostId uuid.UUID
}

type DeletePostResponseDTO struct {
	Success bool
}

type FetchFeedRequestDTO struct {
	NumPosts  int
	Timestamp time.Time
}

type FetchFeedResponseDTO struct {
	Posts []*PostDTO
}

type FetchRecommendationsRequestDTO struct {
	NumPosts  int
	Timestamp time.Time
}

type FetchRecommendationsResponseDTO struct {
	Posts []*PostDTO
}

type FetchUserPostsRequestDTO struct {
	UserId    uuid.UUID
	NumPosts  int
	Timestamp time.Time
}

type FetchUserPostsResponseDTO struct {
	Posts []*PostDTO
}

type UpdatePostRequestDTO struct {
	Post *PostDTO
}

type UpdatePostResponseDTO struct {
	Post *PostDTO
}
