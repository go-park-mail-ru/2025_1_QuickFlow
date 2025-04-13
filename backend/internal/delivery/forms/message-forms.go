package forms

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"

	"quickflow/config"
	"quickflow/internal/models"
)

type GetMessagesForm struct {
	MessagesCount int       `json:"messages_count"`
	Ts            time.Time `json:"ts,omitempty"`
}

func (m *GetMessagesForm) GetParams(values url.Values) error {
	var (
		err         error
		numMessages int64
	)

	if !values.Has("messages_count") {
		return errors.New("messages_count parameter missing")
	}

	numMessages, err = strconv.ParseInt(values.Get("messages_count"), 10, 64)
	if err != nil {
		return errors.New("failed to parse messages_count")
	}

	m.MessagesCount = int(numMessages)

	ts, err := time.Parse(config.TimeStampLayout, values.Get("ts"))
	if err != nil {
		ts = time.Now()
	}
	m.Ts = ts
	return nil
}

type MessageOut struct {
	ID             uuid.UUID `json:"id,omitempty"`
	Text           string    `json:"text"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
	IsRead         bool      `json:"is_read"`
	AttachmentURLs []string  `json:"attachment_urls"`

	Sender PublicUserInfoOut `json:"sender"`
	ChatId uuid.UUID         `json:"chat_id"`
}

func ToMessageOut(message models.Message, info models.PublicUserInfo) MessageOut {
	return MessageOut{
		ID:             message.ID,
		Text:           message.Text,
		CreatedAt:      message.CreatedAt.Format(config.TimeStampLayout),
		UpdatedAt:      message.UpdatedAt.Format(config.TimeStampLayout),
		IsRead:         message.IsRead,
		AttachmentURLs: message.AttachmentURLs,

		Sender: PublicUserInfoToOut(info),
		ChatId: message.ChatID,
	}
}

func ToMessagesOut(messages []models.Message, usersInfo map[uuid.UUID]models.PublicUserInfo) []MessageOut {
	var messagesOut []MessageOut
	for _, message := range messages {
		messagesOut = append(messagesOut, MessageOut{
			ID:             message.ID,
			Text:           message.Text,
			CreatedAt:      message.CreatedAt.Format(config.TimeStampLayout),
			UpdatedAt:      message.UpdatedAt.Format(config.TimeStampLayout),
			IsRead:         message.IsRead,
			AttachmentURLs: message.AttachmentURLs,

			Sender: PublicUserInfoToOut(usersInfo[message.SenderID]),
			ChatId: message.ChatID,
		})
	}
	return messagesOut
}

type MessageForm struct {
	Text            string    `form:"text" json:"text"`
	ChatId          uuid.UUID `form:"chat_id" json:"chat_id,omitempty"`
	AttachmentsUrls []string  `form:"attachment_urls" json:"attachment_urls,omitempty"`
	ReceiverId      uuid.UUID `json:"receiver_id,omitempty"`
	SenderId        uuid.UUID `json:"-"`
}

func (f *MessageForm) ToMessageModel() models.Message {
	return models.Message{
		ID:             uuid.New(),
		Text:           f.Text,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		AttachmentURLs: f.AttachmentsUrls,
		ReceiverID:     f.ReceiverId,
		SenderID:       f.SenderId,
		ChatID:         f.ChatId,
	}
}

type MessageRequest struct {
	Type    string      `json:"type"`
	Payload MessageForm `json:"payload"`
}
