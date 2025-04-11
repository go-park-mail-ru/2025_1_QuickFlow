package forms

import (
	"errors"
	"github.com/google/uuid"
	"net/url"
	"quickflow/config"
	"quickflow/internal/models"
	"strconv"
	"time"
)

type GetChatsForm struct {
	ChatsCount int       `json:"chats_count"`
	Ts         time.Time `json:"ts,omitempty"`
}

type ChatOut struct {
	ID          string      `json:"id"`
	Name        string      `json:"name,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	AvatarURL   string      `json:"avatar_url,omitempty"`
	Type        string      `json:"type"`
	LastMessage *MessageOut `json:"last_message,omitempty"`
}

func (g *GetChatsForm) GetParams(values url.Values) error {
	var (
		err      error
		numChats int64
	)

	if !values.Has("chats_count") {
		return errors.New("chats_count parameter missing")
	}

	numChats, err = strconv.ParseInt(values.Get("chats_count"), 10, 64)
	if err != nil {
		return errors.New("failed to parse chats_count")
	}

	g.ChatsCount = int(numChats)

	ts, err := time.Parse(config.TimeStampLayout, values.Get("ts"))
	if err != nil {
		ts = time.Now()
	}
	g.Ts = ts
	return nil
}

func ToChatsOut(chats []models.Chat) []ChatOut {
	var chatsOut []ChatOut
	var chatType string
	for _, chat := range chats {
		if chat.Type == models.ChatTypePrivate {
			chatType = "private"
		} else if chat.Type == models.ChatTypeGroup {
			chatType = "group"
		} else {
			chatType = "unknown"
		}

		chatOut := ChatOut{
			ID:        chat.ID.String(),
			Name:      chat.Name,
			CreatedAt: chat.CreatedAt,
			UpdatedAt: chat.UpdatedAt,
			AvatarURL: chat.AvatarURL,
			Type:      chatType,
		}
		if chat.LastMessage.ID != uuid.Nil {
			msg := ToMessageOut(chat.LastMessage)
			chatOut.LastMessage = &msg
		}
		chatsOut = append(chatsOut, chatOut)
	}
	return chatsOut
}
