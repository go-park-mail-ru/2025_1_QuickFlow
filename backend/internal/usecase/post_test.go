package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/internal/models"
	mock_usecase "quickflow/internal/usecase/mocks"
)

func TestAddPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockPostRepository(ctrl)
	service := NewPostService(mockRepo)
	ctx := context.Background()

	post := models.Post{
		CreatorId: uuid.New(),
		Desc:      "Test post",
		Images:    []string{"pic1.jpg", "pic2.jpg"},
		CreatedAt: time.Now(),
	}

	mockRepo.EXPECT().AddPost(ctx, gomock.Any()).Return(nil).Times(1)

	err := service.AddPost(ctx, post)
	assert.NoError(t, err)
}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockPostRepository(ctrl)
	service := NewPostService(mockRepo)
	ctx := context.Background()
	postID := uuid.New()

	mockRepo.EXPECT().DeletePost(ctx, postID).Return(nil).Times(1)

	err := service.DeletePost(ctx, postID)
	assert.NoError(t, err)
}

func TestFetchFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockPostRepository(ctrl)
	service := NewPostService(mockRepo)
	ctx := context.Background()
	user := models.User{Id: uuid.New()}
	timestamp := time.Now()
	expectedPosts := []models.Post{
		{Id: uuid.New(), CreatorId: user.Id, Desc: "Post 1", CreatedAt: timestamp},
		{Id: uuid.New(), CreatorId: user.Id, Desc: "Post 2", CreatedAt: timestamp},
	}

	mockRepo.EXPECT().GetPostsForUId(ctx, user.Id, 2, timestamp).Return(expectedPosts, nil).Times(1)

	posts, err := service.FetchFeed(ctx, user, 2, timestamp)
	assert.NoError(t, err)
	assert.Equal(t, expectedPosts, posts)
}

func TestFetchFeed_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_usecase.NewMockPostRepository(ctrl)
	service := NewPostService(mockRepo)
	ctx := context.Background()
	user := models.User{Id: uuid.New()}
	timestamp := time.Now()

	mockRepo.EXPECT().GetPostsForUId(ctx, user.Id, 2, timestamp).Return([]models.Post{}, errors.New("db error")).Times(1)

	posts, err := service.FetchFeed(ctx, user, 2, timestamp)
	assert.Error(t, err)
	assert.Empty(t, posts)
}
