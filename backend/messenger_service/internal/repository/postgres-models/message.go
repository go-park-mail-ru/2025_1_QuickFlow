package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type PostgresFile struct {
	URL         pgtype.Text
	DisplayType pgtype.Text
	Name        pgtype.Text
}

func (f *PostgresFile) ToFile() *models.File {
	file := models.File{
		URL: f.URL.String,
	}
	if f.DisplayType.Valid && len(f.DisplayType.String) != 0 {
		file.DisplayType = models.DisplayType(f.DisplayType.String)
	} else {
		file.DisplayType = models.DisplayTypeFile
	}

	if f.Name.Valid {
		file.Name = f.Name.String
	} else {
		file.Name = ""
	}
	return &file
}

type MessagePostgres struct {
	ID          pgtype.UUID
	Text        pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	Attachments []PostgresFile
	SenderID    pgtype.UUID
	ChatID      pgtype.UUID
}

func (m *MessagePostgres) ToMessage() models.Message {
	var attSlice []*models.File

	for _, att := range m.Attachments {
		if !att.DisplayType.Valid {
			attSlice = append(attSlice, &models.File{URL: att.URL.String, DisplayType: models.DisplayTypeFile, Name: att.Name.String})
		} else {
			attSlice = append(attSlice, &models.File{URL: att.URL.String, DisplayType: models.DisplayType(att.DisplayType.String), Name: att.Name.String})
		}
	}

	return models.Message{
		ID:          m.ID.Bytes,
		Text:        m.Text.String,
		CreatedAt:   m.CreatedAt.Time,
		UpdatedAt:   m.UpdatedAt.Time,
		Attachments: attSlice,
		SenderID:    m.SenderID.Bytes,
		ChatID:      m.ChatID.Bytes,
	}
}

func FromMessage(message models.Message) MessagePostgres {
	var attSlice []PostgresFile

	for _, att := range message.Attachments {
		pgfile := PostgresFile{URL: pgtype.Text{String: att.URL, Valid: true}}
		if len(att.DisplayType) > 0 {
			pgfile.DisplayType = pgtype.Text{String: string(att.DisplayType), Valid: true}
		}
		attSlice = append(attSlice, pgfile)
	}

	return MessagePostgres{
		ID:          pgtype.UUID{Bytes: message.ID, Valid: true},
		Text:        pgtype.Text{String: message.Text, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: message.CreatedAt, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: message.UpdatedAt, Valid: true},
		Attachments: attSlice,
		SenderID:    pgtype.UUID{Bytes: message.SenderID, Valid: true},
		ChatID:      pgtype.UUID{Bytes: message.ChatID, Valid: true},
	}
}
