package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/utils/validation"
)

var (
	ErrInvalidProfileInfo = fmt.Errorf("invalid profile info")
)

type ProfileRepository interface {
	SaveProfile(ctx context.Context, profile models.Profile) error
	GetProfile(ctx context.Context, userId uuid.UUID) (models.Profile, error)
	UpdateProfileTextInfo(ctx context.Context, newProfile models.Profile) error
	UpdateProfileAvatar(ctx context.Context, id uuid.UUID, url string) error
	UpdateProfileCover(ctx context.Context, id uuid.UUID, url string) error
}

type ProfileService struct {
	userRepo    UserRepository
	profileRepo ProfileRepository
	fileRepo    FileRepository
}

// NewProfileService creates new profile service.
func NewProfileService(profileRepo ProfileRepository, userRepo UserRepository, fileRepo FileRepository) *ProfileService {
	return &ProfileService{
		profileRepo: profileRepo,
		fileRepo:    fileRepo,
		userRepo:    userRepo,
	}
}

// GetUserInfo gets user profile information.
func (p *ProfileService) GetUserInfo(ctx context.Context, userId uuid.UUID) (models.Profile, error) {
	profile, err := p.profileRepo.GetProfile(ctx, userId)
	if err != nil {
		return models.Profile{}, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	return profile, nil
}

func (p *ProfileService) GetUserInfoByUserName(ctx context.Context, username string) (models.Profile, error) {
	user, err := p.userRepo.GetUserByUsername(ctx, username)
	if errors.Is(err, ErrNotFound) {
		return models.Profile{}, err
	} else if err != nil {
		return models.Profile{}, fmt.Errorf("p.userRepo.GetUserByUsername: %w", err)
	}

	profile, err := p.profileRepo.GetProfile(ctx, user.Id)
	if err != nil {
		return models.Profile{}, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	return profile, nil
}

// UpdateProfile updates profile in the repository.
func (p *ProfileService) UpdateProfile(ctx context.Context, newProfile models.Profile) error {
	if newProfile.BasicInfo != nil {
		if validation.ValidateProfile(newProfile.BasicInfo.Name, newProfile.BasicInfo.Surname) != nil {
			return ErrInvalidProfileInfo
		}
	}

	err := p.profileRepo.UpdateProfileTextInfo(ctx, newProfile)
	if err != nil {
		return fmt.Errorf("p.profileRepo.UpdateProfileTextInfo: %w", err)
	}

	if newProfile.Avatar != nil {
		avatarUrl, err := p.fileRepo.UploadFile(ctx, newProfile.Avatar)
		if err != nil {
			return fmt.Errorf("p.fileRepo.UploadFile: %w", err)
		}

		err = p.profileRepo.UpdateProfileAvatar(ctx, newProfile.UserId, avatarUrl)
		if err != nil {
			return fmt.Errorf("p.profileRepo.UpdateProfileAvatar: %w", err)
		}
	}

	if newProfile.Background != nil {
		backgroundUrl, err := p.fileRepo.UploadFile(ctx, newProfile.Background)
		if err != nil {
			return fmt.Errorf("p.fileRepo.UploadFile: %w", err)
		}

		err = p.profileRepo.UpdateProfileCover(ctx, newProfile.UserId, backgroundUrl)
		if err != nil {
			return fmt.Errorf("p.profileRepo.UpdateProfileCover: %w", err)
		}
	}

	return nil
}
