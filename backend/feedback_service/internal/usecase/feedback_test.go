package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/feedback_service/internal/usecase"
	"quickflow/feedback_service/internal/usecase/mocks"
	"quickflow/shared/models"
)

func TestSaveFeedback_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFeedbackRepository(ctrl)
	feedback := &models.Feedback{
		Id:           uuid.New(),
		Rating:       4,
		RespondentId: uuid.New(),
		Text:         "This is feedback",
		Type:         models.FeedbackPost,
		CreatedAt:    time.Now(),
	}

	// Set up expectation
	mockRepo.EXPECT().SaveFeedback(context.Background(), feedback).Return(nil).Times(1)

	// Create FeedbackUseCase
	uc := usecase.NewFeedBackUseCase(mockRepo)

	// Call method
	err := uc.SaveFeedback(context.Background(), feedback)

	// Assert the result
	assert.NoError(t, err)
}

func TestGetNumMessagesSent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFeedbackRepository(ctrl)

	// Test data
	userID := uuid.New()
	expectedCount := int64(10)

	// Set up expectation
	mockRepo.EXPECT().GetNumMessagesSent(context.Background(), userID).Return(expectedCount, nil).Times(1)

	// Create FeedbackUseCase
	uc := usecase.NewFeedBackUseCase(mockRepo)

	// Call method
	result, err := uc.GetNumMessagesSent(context.Background(), userID)

	// Assert the result
	assert.NoError(t, err)
	assert.Equal(t, expectedCount, result)
}

func TestGetNumMessagesSent_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFeedbackRepository(ctrl)

	// Test data
	userID := uuid.New()

	// Set up expectation
	mockRepo.EXPECT().GetNumMessagesSent(context.Background(), userID).Return(int64(0), fmt.Errorf("query error")).Times(1)

	// Create FeedbackUseCase
	uc := usecase.NewFeedBackUseCase(mockRepo)

	// Call method
	result, err := uc.GetNumMessagesSent(context.Background(), userID)

	// Assert the result
	assert.Error(t, err)
	assert.Equal(t, int64(0), result)
}
