package forms

import (
	"errors"
	"net/url"
	time2 "quickflow/monolith/config/time"
	models2 "quickflow/monolith/internal/models"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type GetChatsForm struct {
	ChatsCount int       `json:"chats_count"`
	Ts         time.Time `json:"ts,omitempty"`
}

type ChatOut struct {
	ID              string      `json:"id"`
	Name            string      `json:"name,omitempty"`
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"updated_at"`
	AvatarURL       string      `json:"avatar_url,omitempty"`
	Type            string      `json:"type"`
	LastMessage     *MessageOut `json:"last_message,omitempty"`
	IsOnline        *bool       `json:"online,omitempty"`
	LastSeen        string      `json:"last_seen,omitempty"`
	Username        string      `json:"username,omitempty"`
	LastReadByOther string      `json:"last_read_by_other,omitempty"`
	LastReadByMe    string      `json:"last_read_by_me,omitempty"`
}

type PrivateChatInfo struct {
	Username string   `json:"username,omitempty"`
	Activity Activity `json:"activity,omitempty"`
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

	ts, err := time.Parse(time2.TimeStampLayout, values.Get("ts"))
	if err != nil {
		ts = time.Now()
	}
	g.Ts = ts
	return nil
}

func ToChatsOut(chats []models2.Chat, lastMessageSenderInfo map[uuid.UUID]models2.PublicUserInfo, privateChatsOnlineStatus map[uuid.UUID]PrivateChatInfo) []ChatOut {
	var chatsOut []ChatOut
	var chatType string
	for _, chat := range chats {
		if chat.Type == models2.ChatTypePrivate {
			chatType = "private"
		} else if chat.Type == models2.ChatTypeGroup {
			chatType = "group"
		} else {
			chatType = "unknown"
		}

		chatOut := ChatOut{
			ID:        chat.ID.String(),
			Name:      chat.Name,
			CreatedAt: chat.CreatedAt.Format(time2.TimeStampLayout),
			UpdatedAt: chat.UpdatedAt.Format(time2.TimeStampLayout),
			AvatarURL: chat.AvatarURL,
			Type:      chatType,
		}
		if chat.LastReadByOther != nil {
			chatOut.LastReadByOther = chat.LastReadByOther.Format(time2.TimeStampLayout)
		}
		if chat.LastReadByMe != nil {
			chatOut.LastReadByMe = chat.LastReadByMe.Format(time2.TimeStampLayout)
		}
		if chat.LastMessage.ID != uuid.Nil {
			msg := ToMessageOut(chat.LastMessage, lastMessageSenderInfo[chat.LastMessage.SenderID])
			chatOut.LastMessage = &msg
		}

		if chat.Type == models2.ChatTypePrivate && privateChatsOnlineStatus != nil {
			if profileInfo, exists := privateChatsOnlineStatus[chat.ID]; exists {
				chatOut.IsOnline = &profileInfo.Activity.IsOnline
				if !profileInfo.Activity.IsOnline {
					chatOut.LastSeen = profileInfo.Activity.LastSeen
				}
				chatOut.Username = profileInfo.Username
			}
		}
		chatsOut = append(chatsOut, chatOut)
	}
	return chatsOut
}
