package errors

import (
	"errors"
	"fmt"
)

// Error messages
var (
	ErrInvalidNumMessages = fmt.Errorf("numMessages must be greater than 0")
	ErrNotParticipant     = fmt.Errorf("user is not a participant in the chat")
	ErrNotFound           = errors.New("not found")
)

// Error chats
var (
	ErrInvalidChatCreationInfo = fmt.Errorf("invalid chat creation info")
	ErrAlreadyInChat           = fmt.Errorf("user already in chat")
	ErrInvalidChatType         = fmt.Errorf("invalid chat type")
)
