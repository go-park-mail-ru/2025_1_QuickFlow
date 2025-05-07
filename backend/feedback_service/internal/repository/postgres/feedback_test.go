package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestSaveFeedback(t *testing.T) {
	// Initialize mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := NewFeedbackRepository(db)

	feedback := &models.Feedback{
		Id:           uuid.New(),
		Rating:       5,
		RespondentId: uuid.New(),
		Text:         "Great service!",
		Type:         models.FeedbackGeneral,
		CreatedAt:    time.Now(),
	}

	// Mock the SQL execution
	mock.ExpectExec("insert into feedback").
		WithArgs(feedback.Rating, feedback.RespondentId, feedback.Text, feedback.Type, feedback.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method
	err = repo.SaveFeedback(context.Background(), feedback)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveFeedback_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := NewFeedbackRepository(db)

	feedback := &models.Feedback{
		Id:           uuid.New(),
		Rating:       5,
		RespondentId: uuid.New(),
		Text:         "Great service!",
		Type:         models.FeedbackGeneral,
		CreatedAt:    time.Now(),
	}

	// Simulate a SQL error
	mock.ExpectExec("insert into feedback").
		WithArgs(feedback.Rating, feedback.RespondentId, feedback.Text, feedback.Type, feedback.CreatedAt).
		WillReturnError(errors.New("database error"))

	// Call the method
	err = repo.SaveFeedback(context.Background(), feedback)
	assert.Error(t, err)
	assert.Equal(t, "save feedback: database error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllFeedbackType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := NewFeedbackRepository(db)
	feedbackType := models.FeedbackGeneral
	ts := time.Now()
	count := 5

	// Mock query result
	rows := sqlmock.NewRows([]string{"rating", "respondent_id", "text", "type", "created_at"}).
		AddRow(5, uuid.New(), "Great!", feedbackType, time.Now())

	mock.ExpectQuery("select rating, respondent_id, text, type, created_at").
		WithArgs(ts, feedbackType, count).
		WillReturnRows(rows)

	// Call the method
	feedbacks, err := repo.GetAllFeedbackType(context.Background(), models.FeedbackType(feedbackType), ts, count)
	assert.NoError(t, err)
	assert.Len(t, feedbacks, 1)
	assert.Equal(t, feedbacks[0].Rating, 5)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllFeedbackType_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := NewFeedbackRepository(db)
	feedbackType := models.FeedbackGeneral
	ts := time.Now()
	count := 5

	// Simulate a query error
	mock.ExpectQuery("select rating, respondent_id, text, type, created_at").
		WithArgs(ts, feedbackType, count).
		WillReturnError(errors.New("query error"))

	// Call the method
	feedbacks, err := repo.GetAllFeedbackType(context.Background(), models.FeedbackType(feedbackType), ts, count)
	assert.Error(t, err)
	assert.Nil(t, feedbacks)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNumMessagesSent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := NewFeedbackRepository(db)
	userId := uuid.New()

	// Mock query result
	mock.ExpectQuery("select count").
		WithArgs(userId).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	// Call the method
	num, err := repo.GetNumMessagesSent(context.Background(), userId)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), num)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNumMessagesSent_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	repo := NewFeedbackRepository(db)
	userId := uuid.New()

	// Simulate a query error
	mock.ExpectQuery("select count").
		WithArgs(userId).
		WillReturnError(errors.New("query error"))

	// Call the method
	num, err := repo.GetNumMessagesSent(context.Background(), userId)
	assert.Error(t, err)
	assert.Equal(t, int64(0), num)
	assert.NoError(t, mock.ExpectationsWereMet())
}
