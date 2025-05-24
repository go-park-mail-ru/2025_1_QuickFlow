package validation

import (
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"

	"quickflow/community_service/config"
	community_errors "quickflow/community_service/internal/errors"
	"quickflow/shared/models"
)

type CommunityValidator struct {
	communityConfig config.CommunityConfig
}

func NewCommunityValidator(communityConfig config.CommunityConfig) *CommunityValidator {
	return &CommunityValidator{
		communityConfig: communityConfig,
	}
}

func (p *CommunityValidator) ValidateCommunity(community *models.Community) error {
	if community == nil {
		return errors.New("community cannot be nil")
	}
	// TODO
	if len(community.NickName) <= p.communityConfig.CommunityNameMinLength {
		return community_errors.ErrorCommunityNameTooShort
	}
	if len(community.NickName) > p.communityConfig.CommunityNameMaxLength {
		return community_errors.ErrorCommunityNameTooLong
	}

	if community.BasicInfo != nil && utf8.RuneCountInString(community.BasicInfo.Description) > p.communityConfig.CommunityDescriptionMaxLength {
		return community_errors.ErrorCommunityDescriptionTooLong
	}

	if community.Avatar != nil {
		if community.Avatar.Size > p.communityConfig.CommunityAvatarMaxSize {
			return community_errors.ErrorCommunityAvatarSizeExceeded
		}
	}
	if community.OwnerID == uuid.Nil {
		return community_errors.ErrNilOwnerId
	}
	return nil
}
