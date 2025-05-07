package validation

import (
	"errors"

	"quickflow/shared/models"
)

func ValidateChatCreationInfo(chatInfo models.ChatCreationInfo) error {
	switch chatInfo.Type {
	case models.ChatTypePrivate:
		if len(chatInfo.Name) != 0 {
			return errors.New("unexpected name for private chat")
		}
		if chatInfo.Avatar != nil {
			return errors.New("unexpected avatar for private chat")
		}
	case models.ChatTypeGroup:
		if len(chatInfo.Name) == 0 {
			return errors.New("empty name for group chat")
		}
		if len(chatInfo.Name) > 30 {
			return errors.New("name too long for group chat")
		}
		if len(chatInfo.Name) < 3 {
			return errors.New("name too short for group chat")
		}
	default:
		return errors.New("invalid chat type")
	}
	return nil
}
