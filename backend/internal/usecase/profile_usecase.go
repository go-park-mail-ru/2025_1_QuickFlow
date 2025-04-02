package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/utils/validation"
)

type ProfileRepository interface {
	SaveProfile(ctx context.Context, profile models.Profile) error
	GetProfile(ctx context.Context, userId uuid.UUID) (models.Profile, error)
	UpdateProfile(ctx context.Context, newProfile models.Profile) error
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
		return models.Profile{}, fmt.Errorf("p.repo.GetProfile: %w", err)
	}

	return profile, nil
}

func (p *ProfileService) GetUserInfoByUserName(ctx context.Context, username string) (models.Profile, error) {
	user, err := p.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return models.Profile{}, err
	}

	profile, err := p.profileRepo.GetProfile(ctx, user.Id)
	if err != nil {
		return models.Profile{}, err
	}

	return profile, nil
}

// UpdateProfile updates profile in the repository.
func (p *ProfileService) UpdateProfile(ctx context.Context, newProfile models.Profile) error {
	if validation.ValidateProfile(newProfile.Name, newProfile.Surname) != nil {
		return errors.New("invalid profile info")
	}

	if newProfile.Avatar != nil {
		avatarUrl, err := p.fileRepo.UploadFile(ctx, newProfile.Avatar)
		if err != nil {
			return fmt.Errorf("p.fileRepo.UploadFile: %w", err)
		}

		newProfile.AvatarUrl = avatarUrl
	}

	if newProfile.Background != nil {
		backgroundUrl, err := p.fileRepo.UploadFile(ctx, newProfile.Background)
		if err != nil {
			return fmt.Errorf("p.fileRepo.UploadFile: %w", err)
		}

		newProfile.BackgroundUrl = backgroundUrl
	}

	err := p.profileRepo.UpdateProfile(ctx, newProfile)
	if err != nil {
		return fmt.Errorf("p.repo.UpdateProfile: %w", err)
	}

	return nil
}
