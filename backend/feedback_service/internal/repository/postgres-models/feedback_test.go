package postgres_models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

// Функция для сравнения времени с погрешностью в миллисекунды
func compareTimes(t1, t2 time.Time) bool {
	return t1.Truncate(time.Millisecond).Equal(t2.Truncate(time.Millisecond))
}

func TestFromModel(t *testing.T) {
	uuid_ := uuid.New()
	tests := []struct {
		name     string
		feedback models.Feedback
		expected PgFeedback
	}{
		{
			name: "Valid feedback",
			feedback: models.Feedback{
				Id:           uuid_,
				Rating:       4,
				RespondentId: uuid_,
				Text:         "This is feedback",
				Type:         models.FeedbackPost,
				CreatedAt:    time.Now(),
			},
			expected: PgFeedback{
				Id:           pgtype.UUID{Bytes: uuid_, Valid: true},
				Rating:       pgtype.Int4{Int32: 4, Valid: true},
				RespondentId: pgtype.UUID{Bytes: uuid_, Valid: true},
				Text:         pgtype.Text{String: "This is feedback", Valid: true},
				Type:         pgtype.Text{String: "post", Valid: true},
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
		},
		{
			name: "Empty text",
			feedback: models.Feedback{
				Id:           uuid_,
				Rating:       5,
				RespondentId: uuid_,
				Text:         "",
				Type:         models.FeedbackGeneral,
				CreatedAt:    time.Now(),
			},
			expected: PgFeedback{
				Id:           pgtype.UUID{Bytes: uuid_, Valid: true},
				Rating:       pgtype.Int4{Int32: 5, Valid: true},
				RespondentId: pgtype.UUID{Bytes: uuid_, Valid: true},
				Text:         pgtype.Text{Valid: false},
				Type:         pgtype.Text{String: "general", Valid: true},
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromModel(&tt.feedback)
			// Сравниваем только время с погрешностью
			assert.True(t, compareTimes(tt.expected.CreatedAt.Time, result.CreatedAt.Time))
			tt.expected.CreatedAt.Time = result.CreatedAt.Time // Обрезаем время для дальнейших сравнений
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestToModel(t *testing.T) {
	uuid_ := uuid.New()
	tests := []struct {
		name       string
		pgFeedback PgFeedback
		expected   models.Feedback
	}{
		{
			name: "Valid PgFeedback",
			pgFeedback: PgFeedback{
				Id:           pgtype.UUID{Bytes: uuid_, Valid: true},
				Rating:       pgtype.Int4{Int32: 4, Valid: true},
				RespondentId: pgtype.UUID{Bytes: uuid_, Valid: true},
				Text:         pgtype.Text{String: "This is feedback", Valid: true},
				Type:         pgtype.Text{String: "post", Valid: true},
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
			expected: models.Feedback{
				Id:           uuid_,
				Rating:       4,
				RespondentId: uuid_,
				Text:         "This is feedback",
				Type:         models.FeedbackPost,
				CreatedAt:    time.Now(),
			},
		},
		{
			name: "Empty Text in PgFeedback",
			pgFeedback: PgFeedback{
				Id:           pgtype.UUID{Bytes: uuid_, Valid: true},
				Rating:       pgtype.Int4{Int32: 5, Valid: true},
				RespondentId: pgtype.UUID{Bytes: uuid_, Valid: true},
				Text:         pgtype.Text{Valid: false},
				Type:         pgtype.Text{String: "general", Valid: true},
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			},
			expected: models.Feedback{
				Id:           uuid_,
				Rating:       5,
				RespondentId: uuid_,
				Text:         "",
				Type:         models.FeedbackGeneral,
				CreatedAt:    time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pgFeedback.ToModel()
			// Сравниваем только время с погрешностью
			assert.True(t, compareTimes(tt.expected.CreatedAt, result.CreatedAt))
			tt.expected.CreatedAt = result.CreatedAt // Обрезаем время для дальнейших сравнений
			assert.Equal(t, tt.expected, result)
		})
	}
}
