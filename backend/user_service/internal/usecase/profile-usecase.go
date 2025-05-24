package usecase

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	shared_models "quickflow/shared/models"
	user_errors "quickflow/user_service/internal/errors"
	"quickflow/user_service/utils/validation"
)

type ProfileRepository interface {
	SaveProfile(ctx context.Context, profile shared_models.Profile) error
	GetProfile(ctx context.Context, userId uuid.UUID) (shared_models.Profile, error)
	UpdateProfileTextInfo(ctx context.Context, newProfile shared_models.Profile) error
	UpdateProfileAvatar(ctx context.Context, id uuid.UUID, url string) error
	UpdateProfileCover(ctx context.Context, id uuid.UUID, url string) error
	GetPublicUserInfo(ctx context.Context, userId uuid.UUID) (shared_models.PublicUserInfo, error)
	GetPublicUsersInfo(ctx context.Context, userIds []uuid.UUID) ([]shared_models.PublicUserInfo, error)
	UpdateLastSeen(ctx context.Context, userId uuid.UUID) error
}

type FileService interface {
	UploadFile(ctx context.Context, file *shared_models.File) (string, error)
	DeleteFile(ctx context.Context, filename string) error
}

type ProfileService struct {
	userRepo    UserRepository
	profileRepo ProfileRepository
	fileRepo    FileService
}

// NewProfileService creates new profile service.
func NewProfileService(profileRepo ProfileRepository, userRepo UserRepository, fileRepo FileService) *ProfileService {
	return &ProfileService{
		profileRepo: profileRepo,
		fileRepo:    fileRepo,
		userRepo:    userRepo,
	}
}

// GetUserInfo gets user profile information.
func (p *ProfileService) GetUserInfo(ctx context.Context, userId uuid.UUID) (shared_models.Profile, error) {
	profile, err := p.profileRepo.GetProfile(ctx, userId)
	if err != nil {
		return shared_models.Profile{}, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	return profile, nil
}

func (p *ProfileService) GetUserInfoByUserName(ctx context.Context, username string) (shared_models.Profile, error) {
	user, err := p.userRepo.GetUserByUsername(ctx, username)
	if errors.Is(err, user_errors.ErrNotFound) {
		return shared_models.Profile{}, err
	} else if err != nil {
		return shared_models.Profile{}, fmt.Errorf("p.userRepo.GetUserByUsername: %w", err)
	}

	profile, err := p.profileRepo.GetProfile(ctx, user.Id)
	if err != nil {
		return shared_models.Profile{}, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	return profile, nil
}

// UpdateProfile updates profile in the repository.
func (p *ProfileService) UpdateProfile(ctx context.Context, newProfile shared_models.Profile) (*shared_models.Profile, error) {
	if newProfile.BasicInfo != nil {
		if validation.ValidateProfile(newProfile.BasicInfo.Name, newProfile.BasicInfo.Surname) != nil {
			return nil, user_errors.ErrInvalidProfileInfo
		}
	}

	// check if user with this username already exists
	if len(newProfile.Username) != 0 {
		user, err := p.userRepo.GetUserByUsername(ctx, newProfile.Username)
		if err != nil && !errors.Is(err, user_errors.ErrNotFound) {
			return nil, fmt.Errorf("p.userRepo.GetUserByUsername: %w", err)
		}
		if err == nil && user.Id != uuid.Nil && user.Id != newProfile.UserId {
			return nil, user_errors.ErrAlreadyExists
		}
	}

	oldProfile, err := p.profileRepo.GetProfile(ctx, newProfile.UserId)
	if err != nil {
		return nil, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	g, newCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := p.profileRepo.UpdateProfileTextInfo(newCtx, newProfile); err != nil {
			return fmt.Errorf("p.profileRepo.UpdateProfileTextInfo: %w", err)
		}
		return nil
	})

	if newProfile.Avatar != nil {
		g.Go(func() error {
			avatarUrl, err := p.fileRepo.UploadFile(newCtx, newProfile.Avatar)
			if err != nil {
				return fmt.Errorf("p.fileRepo.UploadFile (avatar): %w", err)
			}
			if err := p.profileRepo.UpdateProfileAvatar(newCtx, newProfile.UserId, avatarUrl); err != nil {
				return fmt.Errorf("p.profileRepo.UpdateProfileAvatar: %w", err)
			}
			return nil
		})
	}

	if newProfile.Background != nil {
		g.Go(func() error {
			backgroundUrl, err := p.fileRepo.UploadFile(newCtx, newProfile.Background)
			if err != nil {
				return fmt.Errorf("p.fileRepo.UploadFile (background): %w", err)
			}
			if err := p.profileRepo.UpdateProfileCover(newCtx, newProfile.UserId, backgroundUrl); err != nil {
				return fmt.Errorf("p.profileRepo.UpdateProfileCover: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if newProfile.BasicInfo != nil && oldProfile.BasicInfo != nil {
		if newProfile.Avatar != nil && len(oldProfile.BasicInfo.AvatarUrl) > 0 && oldProfile.BasicInfo.AvatarUrl != newProfile.BasicInfo.AvatarUrl {
			err = p.fileRepo.DeleteFile(ctx, path.Base(oldProfile.BasicInfo.AvatarUrl))
			if err != nil {
				return nil, err
			}
		}
		if newProfile.Background != nil && len(newProfile.BasicInfo.BackgroundUrl) > 0 && newProfile.BasicInfo.BackgroundUrl != newProfile.BasicInfo.BackgroundUrl {
			err = p.fileRepo.DeleteFile(ctx, path.Base(oldProfile.BasicInfo.BackgroundUrl))
			if err != nil {
				return nil, err
			}
		}
	}

	profile, err := p.profileRepo.GetProfile(ctx, newProfile.UserId)
	if err != nil {
		return nil, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	return &profile, nil
}

func (p *ProfileService) GetPublicUserInfo(ctx context.Context, userId uuid.UUID) (shared_models.PublicUserInfo, error) {
	if userId == uuid.Nil {
		return shared_models.PublicUserInfo{}, user_errors.ErrInvalidUserId
	}
	publicInfo, err := p.profileRepo.GetPublicUserInfo(ctx, userId)
	if err != nil {
		return shared_models.PublicUserInfo{}, fmt.Errorf("p.profileRepo.GetPublicUserInfo: %w", err)
	}
	return publicInfo, nil
}

func (p *ProfileService) GetPublicUsersInfo(ctx context.Context, userIds []uuid.UUID) ([]shared_models.PublicUserInfo, error) {
	if len(userIds) == 0 {
		return nil, fmt.Errorf("userIds is empty")
	}

	publicInfo, err := p.profileRepo.GetPublicUsersInfo(ctx, userIds)
	if err != nil {
		return nil, fmt.Errorf("p.profileRepo.GetPublicUsersInfo: %w", err)
	}

	return publicInfo, nil
}

func (p *ProfileService) UpdateLastSeen(ctx context.Context, userId uuid.UUID) error {
	err := p.profileRepo.UpdateLastSeen(ctx, userId)
	if err != nil {
		return fmt.Errorf("a.userRepo.UpdateLastSeen: %w", err)
	}
	return nil
}

func (p *ProfileService) CreateProfile(ctx context.Context, profile shared_models.Profile) (shared_models.Profile, error) {
	if profile.BasicInfo != nil {
		if validation.ValidateProfile(profile.BasicInfo.Name, profile.BasicInfo.Surname) != nil {
			return shared_models.Profile{}, user_errors.ErrInvalidProfileInfo
		}
	}

	if err := p.profileRepo.SaveProfile(ctx, profile); err != nil {
		return shared_models.Profile{}, fmt.Errorf("p.profileRepo.SaveProfile: %w", err)
	}

	return profile, nil
}

func (p *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (shared_models.Profile, error) {
	profile, err := p.profileRepo.GetProfile(ctx, userID)
	if err != nil {
		return shared_models.Profile{}, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	return profile, nil
}

func (p *ProfileService) GetProfileByUsername(ctx context.Context, username string) (shared_models.Profile, error) {
	user, err := p.userRepo.GetUserByUsername(ctx, username)
	if errors.Is(err, user_errors.ErrNotFound) {
		return shared_models.Profile{}, err
	} else if err != nil {
		return shared_models.Profile{}, fmt.Errorf("p.userRepo.GetUserByUsername: %w", err)
	}

	profile, err := p.profileRepo.GetProfile(ctx, user.Id)
	if err != nil {
		return shared_models.Profile{}, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	profile.Username = user.Username

	return profile, nil
}
