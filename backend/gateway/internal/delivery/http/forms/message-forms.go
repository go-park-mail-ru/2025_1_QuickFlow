package forms

import (
    "errors"
    "net/url"
    "strconv"
    "time"

    "github.com/google/uuid"

    time2 "quickflow/config/time"
    "quickflow/shared/models"
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
    ID        uuid.UUID `json:"id,omitempty"`
    Text      string    `json:"text"`
    CreatedAt string    `json:"created_at"`
    UpdatedAt string    `json:"updated_at"`
    MediaURLs []string  `json:"media,omitempty"`
    AudioURLs []string  `json:"audio,omitempty"`
    FileURLs  []string  `json:"files,omitempty"`

    Sender PublicUserInfoOut `json:"sender"`
    ChatId uuid.UUID         `json:"chat_id"`
}

func ToMessageOut(message models.Message, info models.PublicUserInfo) MessageOut {
    mediaURLs := make([]string, 0)
    audioURLs := make([]string, 0)
    fileURLs := make([]string, 0)

    for _, file := range message.Attachments {
        if file.DisplayType == models.DisplayTypeMedia {
            mediaURLs = append(mediaURLs, file.URL)
        } else if file.DisplayType == models.DisplayTypeAudio {
            audioURLs = append(audioURLs, file.URL)
        } else {
            fileURLs = append(fileURLs, file.URL)
        }
    }

    return MessageOut{
        ID:        message.ID,
        Text:      message.Text,
        CreatedAt: message.CreatedAt.Format(time2.TimeStampLayout),
        UpdatedAt: message.UpdatedAt.Format(time2.TimeStampLayout),
        MediaURLs: mediaURLs,
        AudioURLs: audioURLs,
        FileURLs:  fileURLs,

        Sender: PublicUserInfoToOut(info, ""),
        ChatId: message.ChatID,
    }
}

func ToMessagesOut(messages []*models.Message, usersInfo map[uuid.UUID]models.PublicUserInfo) []MessageOut {
    var messagesOut []MessageOut
    for _, message := range messages {
        messagesOut = append(messagesOut, ToMessageOut(*message, usersInfo[message.ChatID]))
    }

    return messagesOut
}

type MessagesOut struct {
    Messages   []MessageOut `json:"messages"`
    LastReadTs string       `json:"last_read_ts,omitempty"`
}

type MessageForm struct {
    Text       string    `form:"text" json:"text,omitempty"`
    ChatId     uuid.UUID `form:"chat_id" json:"chat_id,omitempty"`
    Media      []string  `form:"media" json:"media,omitempty"`
    Audio      []string  `form:"audio" json:"audio,omitempty"`
    File       []string  `form:"files" json:"files,omitempty"`
    ReceiverId uuid.UUID `json:"receiver_id,omitempty"`
    SenderId   uuid.UUID `json:"-"`
}

func (f *MessageForm) ToMessageModel() models.Message {
    var attachments []*models.File
    for _, file := range f.Media {
        attachments = append(attachments, &models.File{
            URL:         file,
            DisplayType: models.DisplayTypeMedia,
        })
    }

    for _, file := range f.Audio {
        attachments = append(attachments, &models.File{
            URL:         file,
            DisplayType: models.DisplayTypeAudio,
        })
    }

    for _, file := range f.File {
        attachments = append(attachments, &models.File{
            URL:         file,
            DisplayType: models.DisplayTypeFile,
        })
    }

    return models.Message{
        ID:          uuid.New(),
        Text:        f.Text,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
        Attachments: attachments,
        ReceiverID:  f.ReceiverId,
        SenderID:    f.SenderId,
        ChatID:      f.ChatId,
    }
}
