package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"
	"quickflow/internal/models"
)

type MessagePostgres struct {
	ID              pgtype.UUID
	Text            pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	IsRead          pgtype.Bool
	AttachmentsURLs []pgtype.Text
	SenderID        pgtype.UUID
	ChatID          pgtype.UUID
}

func (m *MessagePostgres) ToMessage() models.Message {
	var attSlice []string

	for _, att := range m.AttachmentsURLs {
		attSlice = append(attSlice, att.String)
	}

	return models.Message{
		ID:             m.ID.Bytes,
		Text:           m.Text.String,
		CreatedAt:      m.CreatedAt.Time,
		UpdatedAt:      m.UpdatedAt.Time,
		IsRead:         m.IsRead.Bool,
		AttachmentURLs: attSlice,
		SenderID:       m.SenderID.Bytes,
		ChatID:         m.ChatID.Bytes,
	}
}

func FromMessage(message models.Message) MessagePostgres {
	var attSlice []pgtype.Text

	for _, att := range message.AttachmentURLs {
		attSlice = append(attSlice, pgtype.Text{String: att, Valid: true})
	}

	return MessagePostgres{
		ID:              pgtype.UUID{Bytes: message.ID, Valid: true},
		Text:            pgtype.Text{String: message.Text, Valid: true},
		CreatedAt:       pgtype.Timestamptz{Time: message.CreatedAt, Valid: true},
		UpdatedAt:       pgtype.Timestamptz{Time: message.UpdatedAt, Valid: true},
		IsRead:          pgtype.Bool{Bool: message.IsRead, Valid: true},
		AttachmentsURLs: attSlice,
		SenderID:        pgtype.UUID{Bytes: message.SenderID, Valid: true},
		ChatID:          pgtype.UUID{Bytes: message.ChatID, Valid: true},
	}
}
