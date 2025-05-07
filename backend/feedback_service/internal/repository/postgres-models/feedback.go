package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type PgFeedback struct {
	Id           pgtype.UUID
	Rating       pgtype.Int4
	RespondentId pgtype.UUID
	Text         pgtype.Text
	Type         pgtype.Text
	CreatedAt    pgtype.Timestamptz
}

func FromModel(feedback *models.Feedback) *PgFeedback {

	pgFeedback := &PgFeedback{
		Id:           pgtype.UUID{Bytes: feedback.Id, Valid: true},
		Rating:       pgtype.Int4{Int32: int32(feedback.Rating), Valid: true},
		Type:         pgtype.Text{String: string(feedback.Type), Valid: true},
		RespondentId: pgtype.UUID{Bytes: feedback.RespondentId, Valid: true},
		CreatedAt:    pgtype.Timestamptz{Time: feedback.CreatedAt, Valid: true},
	}

	if len(feedback.Text) > 0 {
		pgFeedback.Text = pgtype.Text{String: feedback.Text, Valid: true}
	}

	return pgFeedback
}

func (pf *PgFeedback) ToModel() models.Feedback {
	return models.Feedback{
		Id:           pf.Id.Bytes,
		Rating:       int(pf.Rating.Int32),
		Text:         pf.Text.String,
		RespondentId: pf.RespondentId.Bytes,
		Type:         models.FeedbackType(pf.Type.String),
		CreatedAt:    pf.CreatedAt.Time,
	}
}
