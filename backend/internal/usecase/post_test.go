package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"quickflow/internal/models"
)

// Mock для PostRepository
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) AddPost(ctx context.Context, post models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) DeletePost(ctx context.Context, postId uuid.UUID) error {
	args := m.Called(ctx, postId)
	return args.Error(0)
}

func (m *MockPostRepository) GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
	args := m.Called(ctx, uid, numPosts, timestamp)
	return args.Get(0).([]models.Post), args.Error(1)
}

func TestAddPost(t *testing.T) {
	mockRepo := new(MockPostRepository)
	service := NewPostService(mockRepo)
	ctx := context.Background()

	post := models.Post{
		CreatorId: uuid.New(),
		Desc:      "Test post",
		Pics:      []string{"pic1.jpg", "pic2.jpg"},
		CreatedAt: time.Now(),
	}

	mockRepo.On("AddPost", ctx, mock.Anything).Return(nil).Once()

	err := service.AddPost(ctx, post)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeletePost(t *testing.T) {
	mockRepo := new(MockPostRepository)
	service := NewPostService(mockRepo)
	ctx := context.Background()
	postID := uuid.New()

	mockRepo.On("DeletePost", ctx, postID).Return(nil).Once()

	err := service.DeletePost(ctx, postID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFetchFeed(t *testing.T) {
	mockRepo := new(MockPostRepository)
	service := NewPostService(mockRepo)
	ctx := context.Background()
	user := models.User{Id: uuid.New()}
	timestamp := time.Now()
	expectedPosts := []models.Post{
		{Id: uuid.New(), CreatorId: user.Id, Desc: "Post 1", CreatedAt: timestamp},
		{Id: uuid.New(), CreatorId: user.Id, Desc: "Post 2", CreatedAt: timestamp},
	}

	mockRepo.On("GetPostsForUId", ctx, user.Id, 2, timestamp).Return(expectedPosts, nil).Once()

	posts, err := service.FetchFeed(ctx, user, 2, timestamp)
	assert.NoError(t, err)
	assert.Equal(t, expectedPosts, posts)
	mockRepo.AssertExpectations(t)
}

func TestFetchFeed_Error(t *testing.T) {
	mockRepo := new(MockPostRepository)
	service := NewPostService(mockRepo)
	ctx := context.Background()
	user := models.User{Id: uuid.New()}
	timestamp := time.Now()

	mockRepo.On("GetPostsForUId", ctx, user.Id, 2, timestamp).Return([]models.Post{}, errors.New("db error")).Once()

	posts, err := service.FetchFeed(ctx, user, 2, timestamp)
	assert.Error(t, err)
	assert.Empty(t, posts)
	mockRepo.AssertExpectations(t)
}
