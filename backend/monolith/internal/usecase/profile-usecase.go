package usecase

import (
	"context"
	"errors"
	"fmt"
	models2 "quickflow/monolith/internal/models"
	"quickflow/monolith/utils/validation"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

var (
	ErrInvalidProfileInfo = fmt.Errorf("invalid profile info")
	ErrInvalidUserId      = fmt.Errorf("invalid user id")
)

type ProfileRepository interface {
	SaveProfile(ctx context.Context, profile models2.Profile) error
	GetProfile(ctx context.Context, userId uuid.UUID) (models2.Profile, error)
	UpdateProfileTextInfo(ctx context.Context, newProfile models2.Profile) error
	UpdateProfileAvatar(ctx context.Context, id uuid.UUID, url string) error
	UpdateProfileCover(ctx context.Context, id uuid.UUID, url string) error
	GetPublicUserInfo(ctx context.Context, userId uuid.UUID) (models2.PublicUserInfo, error)
	GetPublicUsersInfo(ctx context.Context, userIds []uuid.UUID) ([]models2.PublicUserInfo, error)
	UpdateLastSeen(ctx context.Context, userId uuid.UUID) error
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
func (p *ProfileService) GetUserInfo(ctx context.Context, userId uuid.UUID) (models2.Profile, error) {
	profile, err := p.profileRepo.GetProfile(ctx, userId)
	if err != nil {
		return models2.Profile{}, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	return profile, nil
}

func (p *ProfileService) GetUserInfoByUserName(ctx context.Context, username string) (models2.Profile, error) {
	user, err := p.userRepo.GetUserByUsername(ctx, username)
	if errors.Is(err, ErrNotFound) {
		return models2.Profile{}, err
	} else if err != nil {
		return models2.Profile{}, fmt.Errorf("p.userRepo.GetUserByUsername: %w", err)
	}

	profile, err := p.profileRepo.GetProfile(ctx, user.Id)
	if err != nil {
		return models2.Profile{}, fmt.Errorf("p.profileRepo.GetProfile: %w", err)
	}

	return profile, nil
}

// UpdateProfile updates profile in the repository.
func (p *ProfileService) UpdateProfile(ctx context.Context, newProfile models2.Profile) error {
	if newProfile.BasicInfo != nil {
		if validation.ValidateProfile(newProfile.BasicInfo.Name, newProfile.BasicInfo.Surname) != nil {
			return ErrInvalidProfileInfo
		}
	}

	// check if user with this username already exists
	if len(newProfile.Username) != 0 {
		user, err := p.userRepo.GetUserByUsername(ctx, newProfile.Username)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("p.userRepo.GetUserByUsername: %w", err)
		}
		if err == nil && user.Id != uuid.Nil && user.Id != newProfile.UserId {
			return ErrAlreadyExists
		}
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := p.profileRepo.UpdateProfileTextInfo(ctx, newProfile); err != nil {
			return fmt.Errorf("p.profileRepo.UpdateProfileTextInfo: %w", err)
		}
		return nil
	})

	if newProfile.Avatar != nil {
		g.Go(func() error {
			avatarUrl, err := p.fileRepo.UploadFile(ctx, newProfile.Avatar)
			if err != nil {
				return fmt.Errorf("p.fileRepo.UploadFile (avatar): %w", err)
			}
			if err := p.profileRepo.UpdateProfileAvatar(ctx, newProfile.UserId, avatarUrl); err != nil {
				return fmt.Errorf("p.profileRepo.UpdateProfileAvatar: %w", err)
			}
			return nil
		})
	}

	if newProfile.Background != nil {
		g.Go(func() error {
			backgroundUrl, err := p.fileRepo.UploadFile(ctx, newProfile.Background)
			if err != nil {
				return fmt.Errorf("p.fileRepo.UploadFile (background): %w", err)
			}
			if err := p.profileRepo.UpdateProfileCover(ctx, newProfile.UserId, backgroundUrl); err != nil {
				return fmt.Errorf("p.profileRepo.UpdateProfileCover: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (p *ProfileService) GetPublicUserInfo(ctx context.Context, userId uuid.UUID) (models2.PublicUserInfo, error) {
	if userId == uuid.Nil {
		return models2.PublicUserInfo{}, ErrInvalidUserId
	}
	publicInfo, err := p.profileRepo.GetPublicUserInfo(ctx, userId)
	if err != nil {
		return models2.PublicUserInfo{}, fmt.Errorf("p.profileRepo.GetPublicUserInfo: %w", err)
	}
	return publicInfo, nil
}

func (p *ProfileService) GetPublicUsersInfo(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]models2.PublicUserInfo, error) {
	if len(userIds) == 0 {
		return nil, fmt.Errorf("userIds is empty")
	}

	publicInfo, err := p.profileRepo.GetPublicUsersInfo(ctx, userIds)
	if err != nil {
		return nil, fmt.Errorf("p.profileRepo.GetPublicUsersInfo: %w", err)
	}

	userInfoMap := make(map[uuid.UUID]models2.PublicUserInfo)
	for _, userInfo := range publicInfo {
		userInfoMap[userInfo.Id] = userInfo
	}
	return userInfoMap, nil
}

func (p *ProfileService) UpdateLastSeen(ctx context.Context, userId uuid.UUID) error {
	err := p.profileRepo.UpdateLastSeen(ctx, userId)
	if err != nil {
		return fmt.Errorf("a.userRepo.UpdateLastSeen: %w", err)
	}
	return nil
}
