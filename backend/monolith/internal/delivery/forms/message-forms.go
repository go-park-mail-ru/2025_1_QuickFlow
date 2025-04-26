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

	ts, err := time.Parse(time2.TimeStampLayout, values.Get("ts"))
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
	AttachmentURLs []string  `json:"attachment_urls"`

	Sender PublicUserInfoOut `json:"sender"`
	ChatId uuid.UUID         `json:"chat_id"`
}

func ToMessageOut(message models2.Message, info models2.PublicUserInfo) MessageOut {
	return MessageOut{
		ID:             message.ID,
		Text:           message.Text,
		CreatedAt:      message.CreatedAt.Format(time2.TimeStampLayout),
		UpdatedAt:      message.UpdatedAt.Format(time2.TimeStampLayout),
		AttachmentURLs: message.AttachmentURLs,

		Sender: PublicUserInfoToOut(info, ""),
		ChatId: message.ChatID,
	}
}

func ToMessagesOut(messages []models2.Message, usersInfo map[uuid.UUID]models2.PublicUserInfo) []MessageOut {
	var messagesOut []MessageOut
	for _, message := range messages {
		messagesOut = append(messagesOut, MessageOut{
			ID:             message.ID,
			Text:           message.Text,
			CreatedAt:      message.CreatedAt.Format(time2.TimeStampLayout),
			UpdatedAt:      message.UpdatedAt.Format(time2.TimeStampLayout),
			AttachmentURLs: message.AttachmentURLs,

			Sender: PublicUserInfoToOut(usersInfo[message.SenderID], ""),
			ChatId: message.ChatID,
		})
	}
	return messagesOut
}

type MessagesOut struct {
	Messages   []MessageOut `json:"messages"`
	LastReadTs string       `json:"last_read_ts,omitempty"`
}

type MessageForm struct {
	Text            string    `form:"text" json:"text"`
	ChatId          uuid.UUID `form:"chat_id" json:"chat_id,omitempty"`
	AttachmentsUrls []string  `form:"attachment_urls" json:"attachment_urls,omitempty"`
	ReceiverId      uuid.UUID `json:"receiver_id,omitempty"`
	SenderId        uuid.UUID `json:"-"`
}

func (f *MessageForm) ToMessageModel() models2.Message {
	return models2.Message{
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
