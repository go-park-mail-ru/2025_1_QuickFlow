package forms_test

import (
    "errors"
    "net/url"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"

    "quickflow/config/time"
    "quickflow/gateway/internal/delivery/forms"
    "quickflow/shared/models"
)

func TestGetMessagesForm_GetParams(t *testing.T) {
    parse, i := time.Parse(time_config.TimeStampLayout, "2023-01-01T00:00:00Z")

    if i != nil {
        return
    }

    tests := []struct {
        name        string
        values      url.Values
        expected    forms.GetMessagesForm
        expectedErr error
    }{
        {
            name: "Success - all fields present",
            values: url.Values{
                "messages_count": []string{"10"},
                "ts":             []string{"2023-01-01T00:00:00Z"},
            },
            expected: forms.GetMessagesForm{
                MessagesCount: 10,
                Ts:            parse,
            },
            expectedErr: nil,
        },
        {
            name: "Missing messages_count",
            values: url.Values{
                "ts": []string{"2023-01-01T00:00:00Z"},
            },
            expected:    forms.GetMessagesForm{},
            expectedErr: errors.New("messages_count parameter missing"),
        },
        {
            name: "Invalid messages_count format",
            values: url.Values{
                "messages_count": []string{"notanumber"},
                "ts":             []string{"2023-01-01T00:00:00Z"},
            },
            expected:    forms.GetMessagesForm{},
            expectedErr: errors.New("failed to parse messages_count"),
        },
        {
            name: "Missing timestamp - uses current time",
            values: url.Values{
                "messages_count": []string{"10"},
            },
            expected: forms.GetMessagesForm{
                MessagesCount: 10,
                Ts:            time.Now(), // Will be compared separately
            },
            expectedErr: nil,
        },
        {
            name: "Invalid timestamp format - uses current time",
            values: url.Values{
                "messages_count": []string{"10"},
                "ts":             []string{"invalid"},
            },
            expected: forms.GetMessagesForm{
                MessagesCount: 10,
                Ts:            time.Now(), // Will be compared separately
            },
            expectedErr: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var form forms.GetMessagesForm
            err := form.GetParams(tt.values)

            if tt.expectedErr != nil {
                assert.EqualError(t, err, tt.expectedErr.Error())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected.MessagesCount, form.MessagesCount)

                // Special handling for timestamp comparison when it should be "now"
                if tt.values.Get("ts") == "" || tt.values.Get("ts") == "invalid" {
                    assert.WithinDuration(t, time.Now(), form.Ts, time.Second)
                } else {
                    assert.Equal(t, tt.expected.Ts, form.Ts)
                }
            }
        })
    }
}

func TestToMessageOut(t *testing.T) {
    now := time.Now()
    msgID := uuid.New()
    userID := uuid.New()
    chatID := uuid.New()

    message := models.Message{
        ID:             msgID,
        Text:           "Hello world",
        CreatedAt:      now,
        UpdatedAt:      now,
        AttachmentURLs: []string{"url1", "url2"},
        SenderID:       userID,
        ChatID:         chatID,
    }

    userInfo := models.PublicUserInfo{
        Id:        userID,
        Username:  "testuser",
        Firstname: "Test",
        Lastname:  "User",
        AvatarURL: "avatar.jpg",
        LastSeen:  now,
    }

    expected := forms.MessageOut{
        ID:             msgID,
        Text:           "Hello world",
        CreatedAt:      now.Format(time_config.TimeStampLayout),
        UpdatedAt:      now.Format(time_config.TimeStampLayout),
        AttachmentURLs: []string{"url1", "url2"},
        Sender: forms.PublicUserInfoOut{
            ID:        userID.String(),
            Username:  "testuser",
            FirstName: "Test",
            LastName:  "User",
            AvatarURL: "avatar.jpg",
        },
        ChatId: chatID,
    }

    result := forms.ToMessageOut(message, userInfo)
    assert.Equal(t, expected, result)
}

func TestToMessagesOut(t *testing.T) {
    now := time.Now()
    msg1ID := uuid.New()
    msg2ID := uuid.New()
    user1ID := uuid.New()
    user2ID := uuid.New()
    chatID := uuid.New()

    messages := []*models.Message{
        {
            ID:             msg1ID,
            Text:           "Message 1",
            CreatedAt:      now,
            UpdatedAt:      now,
            AttachmentURLs: []string{"url1"},
            SenderID:       user1ID,
            ChatID:         chatID,
        },
        {
            ID:             msg2ID,
            Text:           "Message 2",
            CreatedAt:      now.Add(time.Minute),
            UpdatedAt:      now.Add(time.Minute),
            AttachmentURLs: []string{"url2"},
            SenderID:       user2ID,
            ChatID:         chatID,
        },
    }

    usersInfo := map[uuid.UUID]models.PublicUserInfo{
        user1ID: {
            Id:        user1ID,
            Username:  "user1",
            Firstname: "User",
            Lastname:  "One",
            AvatarURL: "avatar1.jpg",
        },
        user2ID: {
            Id:        user2ID,
            Username:  "user2",
            Firstname: "User",
            Lastname:  "Two",
            AvatarURL: "avatar2.jpg",
        },
    }

    expected := []forms.MessageOut{
        {
            ID:             msg1ID,
            Text:           "Message 1",
            CreatedAt:      now.Format(time_config.TimeStampLayout),
            UpdatedAt:      now.Format(time_config.TimeStampLayout),
            AttachmentURLs: []string{"url1"},
            Sender: forms.PublicUserInfoOut{
                ID:        user1ID.String(),
                Username:  "user1",
                FirstName: "User",
                LastName:  "One",
                AvatarURL: "avatar1.jpg",
            },
            ChatId: chatID,
        },
        {
            ID:             msg2ID,
            Text:           "Message 2",
            CreatedAt:      now.Add(time.Minute).Format(time_config.TimeStampLayout),
            UpdatedAt:      now.Add(time.Minute).Format(time_config.TimeStampLayout),
            AttachmentURLs: []string{"url2"},
            Sender: forms.PublicUserInfoOut{
                ID:        user2ID.String(),
                Username:  "user2",
                FirstName: "User",
                LastName:  "Two",
                AvatarURL: "avatar2.jpg",
            },
            ChatId: chatID,
        },
    }

    result := forms.ToMessagesOut(messages, usersInfo)
    assert.Equal(t, expected, result)
}

func TestMessageForm_ToMessageModel(t *testing.T) {
    now := time.Now()
    chatID := uuid.New()
    receiverID := uuid.New()
    senderID := uuid.New()

    form := forms.MessageForm{
        Text:            "Test message",
        ChatId:          chatID,
        AttachmentsUrls: []string{"url1", "url2"},
        ReceiverId:      receiverID,
        SenderId:        senderID,
    }

    result := form.ToMessageModel()

    assert.NotEqual(t, uuid.Nil, result.ID)
    assert.Equal(t, "Test message", result.Text)
    assert.WithinDuration(t, now, result.CreatedAt, time.Second)
    assert.WithinDuration(t, now, result.UpdatedAt, time.Second)
    assert.Equal(t, []string{"url1", "url2"}, result.AttachmentURLs)
    assert.Equal(t, senderID, result.SenderID)
    assert.Equal(t, receiverID, result.ReceiverID)
    assert.Equal(t, chatID, result.ChatID)
}
