package usecase

import (
    "context"
    "github.com/google/uuid"
    "quickflow/internal/models"
)

type ProfileRepository interface {
    SaveProfile(ctx context.Context, profile models.Profile) error
    GetProfile(ctx context.Context, userId uuid.UUID) (models.Profile, error)
    UpdateProfile(ctx context.Context, newProfile models.Profile) error
}

type ProfileService struct {
    profileRepo ProfileRepository
}

// NewProfileService creates new profile service.
func NewProfileService(profileRepo ProfileRepository) *AuthService {
    return &AuthService{
        profileRepo: profileRepo,
    }
}

// GetUserInfo returns user profile info.
func (p *ProfileService) GetUserInfo(ctx context.Context, userId uuid.UUID) (models.Profile, error) {
    profile, err := p.profileRepo.GetProfile(ctx, userId)
    if err != nil {
        return models.Profile{}, err
    }

    return profile, nil
}

func (p *ProfileService) UpdateProfile(ctx context.Context, newProfile models.Profile) error {
    // TODO validation

    err := p.profileRepo.UpdateProfile(ctx, newProfile)
    if err != nil {
        return err
    }

    return nil
}
