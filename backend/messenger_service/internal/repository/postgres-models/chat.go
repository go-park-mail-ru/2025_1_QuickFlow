package postgres_models

import (
	"github.com/jackc/pgx/v5/pgtype"

	"quickflow/shared/models"
)

type ChatPostgres struct {
	Id              pgtype.UUID
	Name            pgtype.Text
	AvatarURL       pgtype.Text
	Type            pgtype.Int4
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	LastReadByOther pgtype.Timestamptz
	LastReadByMe    pgtype.Timestamptz
	Messages        []MessagePostgres
}

func (c *ChatPostgres) ToChat() *models.Chat {
	chat := &models.Chat{
		ID:        c.Id.Bytes,
		Name:      getStringIfValid(c.Name),
		AvatarURL: getStringIfValid(c.AvatarURL),
		Type:      models.ChatType(c.Type.Int32),
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}
	if c.LastReadByOther.Valid {
		tm := c.LastReadByOther.Time
		chat.LastReadByOther = &tm
	}
	if c.LastReadByMe.Valid {
		tm := c.LastReadByMe.Time
		chat.LastReadByMe = &tm
	}
	return chat
}

func ModelToPostgres(chat *models.Chat) *ChatPostgres {
	chatPostgres := &ChatPostgres{
		Id:        pgtype.UUID{Bytes: chat.ID, Valid: true},
		Type:      pgtype.Int4{Int32: int32(chat.Type), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: chat.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: chat.UpdatedAt, Valid: true},
	}

	if chat.LastReadByOther != nil {
		chatPostgres.LastReadByOther = pgtype.Timestamptz{Time: *chat.LastReadByOther, Valid: true}
	}
	if len(chat.Name) == 0 {
		chatPostgres.Name = pgtype.Text{Valid: false}
	} else {
		chatPostgres.Name = pgtype.Text{String: chat.Name, Valid: true}
	}

	if len(chat.AvatarURL) == 0 {
		chatPostgres.AvatarURL = pgtype.Text{Valid: false}
	} else {
		chatPostgres.AvatarURL = pgtype.Text{String: chat.AvatarURL, Valid: true}
	}
	return chatPostgres
}

func getStringIfValid(s pgtype.Text) string {
	if s.Valid {
		return s.String
	}
	return ""
}
