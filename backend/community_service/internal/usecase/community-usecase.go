package usecase

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/google/uuid"

	community_errors "quickflow/community_service/internal/errors"
	"quickflow/shared/logger"
	"quickflow/shared/models"
)

type CommunityRepository interface {
	CreateCommunity(ctx context.Context, community models.Community) error
	GetCommunityById(ctx context.Context, id uuid.UUID) (models.Community, error)
	GetCommunityByName(ctx context.Context, name string) (models.Community, error)
	GetCommunityMembers(ctx context.Context, id uuid.UUID, numMembers int, ts time.Time) ([]models.CommunityMember, error)
	IsCommunityMember(ctx context.Context, userId, communityId uuid.UUID) (bool, *models.CommunityRole, error)
	DeleteCommunity(ctx context.Context, id uuid.UUID) error
	UpdateCommunityTextInfo(ctx context.Context, community models.Community) error
	UpdateCommunityAvatar(ctx context.Context, communityId uuid.UUID, avatarUrl string) error
	UpdateCommunityCover(ctx context.Context, communityId uuid.UUID, coverUrl string) error
	JoinCommunity(ctx context.Context, member models.CommunityMember) error
	LeaveCommunity(ctx context.Context, userId, communityId uuid.UUID) error
	GetUserCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]models.Community, error)
	SearchSimilarCommunities(ctx context.Context, name string, count int) ([]models.Community, error)
	ChangeUserRole(ctx context.Context, userId, communityId uuid.UUID, role models.CommunityRole) error
	GetControlledCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]models.Community, error)
}

type CommunityValidator interface {
	ValidateCommunity(community *models.Community) error
}

type FileService interface {
	UploadFile(ctx context.Context, file *models.File) (string, error)
	DeleteFile(ctx context.Context, fileUrl string) error
}

type CommunityUseCase struct {
	repo        CommunityRepository
	fileService FileService
	validator   CommunityValidator
}

func NewCommunityUseCase(repo CommunityRepository, fileService FileService, validator CommunityValidator) *CommunityUseCase {
	return &CommunityUseCase{
		repo:        repo,
		fileService: fileService,
		validator:   validator,
	}
}

func (c *CommunityUseCase) CreateCommunity(ctx context.Context, community models.Community) (*models.Community, error) {
	if err := c.validator.ValidateCommunity(&community); err != nil {
		logger.Error(ctx, fmt.Sprintf("community validation error: %v", err))
		return nil, err
	}

	if community.BasicInfo == nil {
		logger.Error(ctx, "community basic info cannot be nil")
		return nil, fmt.Errorf("community basic info cannot be nil")
	}

	// check if community with this name already exists
	_, err := c.repo.GetCommunityByName(ctx, community.BasicInfo.Name)
	if err != nil && !errors.Is(err, community_errors.ErrNotFound) {
		logger.Error(ctx, fmt.Sprintf("failed to check if community exists: %v", err))
		return nil, err
	}
	if err == nil {
		logger.Info(ctx, "community with this name already exists")
		return nil, community_errors.ErrAlreadyExists
	}

	if community.ID == uuid.Nil {
		community.ID = uuid.New()
	}
	if community.CreatedAt.IsZero() {
		community.CreatedAt = time.Now()
	}

	// upload file if necessary
	if community.Avatar != nil {
		community.BasicInfo.AvatarUrl, err = c.fileService.UploadFile(ctx, community.Avatar)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to upload community avatar: %v", err))
			return nil, err
		}
	}

	if community.Cover != nil {
		community.BasicInfo.CoverUrl, err = c.fileService.UploadFile(ctx, community.Cover)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to upload community cover: %v", err))
			return nil, err
		}
	}

	if err = c.repo.CreateCommunity(ctx, community); err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to create community: %v", err))
		return nil, err
	}
	return &community, nil
}

func (c *CommunityUseCase) GetCommunityById(ctx context.Context, id uuid.UUID) (models.Community, error) {
	if id == uuid.Nil {
		logger.Error(ctx, "community ID cannot be empty")
		return models.Community{}, fmt.Errorf("community ID cannot be empty")
	}

	community, err := c.repo.GetCommunityById(ctx, id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get community by ID: %v", err))
		return models.Community{}, err
	}
	return community, nil
}

func (c *CommunityUseCase) GetCommunityByName(ctx context.Context, name string) (models.Community, error) {
	if len(name) == 0 {
		logger.Error(ctx, "community name cannot be empty")
		return models.Community{}, fmt.Errorf("community name cannot be empty")
	}

	community, err := c.repo.GetCommunityByName(ctx, name)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get community by name: %v", err))
		return models.Community{}, err
	}
	return community, nil
}

func (c *CommunityUseCase) GetCommunityMembers(ctx context.Context, id uuid.UUID, numMembers int, ts time.Time) ([]models.CommunityMember, error) {
	if id == uuid.Nil {
		logger.Error(ctx, "community ID cannot be empty")
		return nil, fmt.Errorf("community ID cannot be empty")
	}

	members, err := c.repo.GetCommunityMembers(ctx, id, numMembers, ts)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get community members: %v", err))
		return nil, err
	}
	return members, nil
}

func (c *CommunityUseCase) IsCommunityMember(ctx context.Context, userId, communityId uuid.UUID) (bool, *models.CommunityRole, error) {
	if userId == uuid.Nil || communityId == uuid.Nil {
		logger.Error(ctx, "user ID and community ID cannot be empty")
		return false, nil, fmt.Errorf("user ID and community ID cannot be empty")
	}

	var role *models.CommunityRole
	isMember, role, err := c.repo.IsCommunityMember(ctx, userId, communityId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to check if user is a community member: %v", err))
		return false, nil, err
	}
	return isMember, role, nil
}

func (c *CommunityUseCase) DeleteCommunity(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		logger.Error(ctx, "community ID cannot be empty")
		return fmt.Errorf("community ID cannot be empty")
	}

	// get url of photo to delete
	community, err := c.repo.GetCommunityById(ctx, id)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get community by ID: %v", err))
		return err
	}

	if err := c.repo.DeleteCommunity(ctx, id); err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to delete community: %v", err))
		return err
	}

	// delete file if necessary
	if len(community.BasicInfo.AvatarUrl) > 0 {
		err = c.fileService.DeleteFile(ctx, path.Base(community.BasicInfo.AvatarUrl))
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to delete community avatar: %v", err))
			return err
		}
	}

	if len(community.BasicInfo.CoverUrl) > 0 {
		err = c.fileService.DeleteFile(ctx, path.Base(community.BasicInfo.CoverUrl))
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to delete community cover: %v", err))
			return err
		}
	}
	return nil
}

func (c *CommunityUseCase) UpdateCommunity(ctx context.Context, community models.Community, userId uuid.UUID) (*models.Community, error) {
	if err := c.validator.ValidateCommunity(&community); err != nil && !errors.Is(err, community_errors.ErrNilOwnerId) {
		logger.Error(ctx, fmt.Sprintf("community validation error: %v", err))
		return nil, err
	}

	if community.ID == uuid.Nil {
		logger.Error(ctx, "community ID cannot be empty")
		return nil, fmt.Errorf("community ID cannot be empty")
	}

	// check if community with this name already exists
	existingCommunity, err := c.repo.GetCommunityByName(ctx, community.BasicInfo.Name)
	if err != nil && !errors.Is(err, community_errors.ErrNotFound) {
		logger.Error(ctx, fmt.Sprintf("failed to check if community exists: %v", err))
		return nil, err
	}
	if err == nil && existingCommunity.ID != community.ID {
		logger.Info(ctx, "community with this name already exists")
		return nil, community_errors.ErrAlreadyExists
	}

	// get user role
	isMember, role, err := c.repo.IsCommunityMember(ctx, userId, community.ID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to check if user is a community member: %v", err))
		return nil, err
	}

	if !isMember || role == nil {
		logger.Error(ctx, "user is not a member of the community")
		return nil, community_errors.ErrNotParticipant
	}

	if *role != models.CommunityRoleAdmin && *role != models.CommunityRoleOwner {
		return nil, community_errors.ErrForbidden
	}

	// get and delete old picture
	oldCommunity, err := c.repo.GetCommunityById(ctx, community.ID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get community by ID: %v", err))
		return nil, err
	}

	if err := c.repo.UpdateCommunityTextInfo(ctx, community); err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to update community: %v", err))
		return nil, err
	}

	// upload new picture if necessary
	if community.Avatar != nil {
		community.BasicInfo.AvatarUrl, err = c.fileService.UploadFile(ctx, community.Avatar)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to upload community avatar: %v", err))
			return nil, err
		}

		if err := c.repo.UpdateCommunityAvatar(ctx, community.ID, community.BasicInfo.AvatarUrl); err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to update community avatar: %v", err))
			return nil, err
		}
	}

	if community.Cover != nil {
		community.BasicInfo.CoverUrl, err = c.fileService.UploadFile(ctx, community.Cover)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to upload community cover: %v", err))
			return nil, err
		}
		if err := c.repo.UpdateCommunityCover(ctx, community.ID, community.BasicInfo.CoverUrl); err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to update community cover: %v", err))
			return nil, err
		}
	}

	// return updated community
	updatedCommunity, err := c.repo.GetCommunityById(ctx, community.ID)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get updated community by ID: %v", err))
		return nil, err
	}

	if len(oldCommunity.BasicInfo.AvatarUrl) > 0 && oldCommunity.BasicInfo.AvatarUrl != community.BasicInfo.AvatarUrl {
		err = c.fileService.DeleteFile(ctx, path.Base(oldCommunity.BasicInfo.AvatarUrl))
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to delete community avatar: %v", err))
			return nil, err
		}
	}
	if len(oldCommunity.BasicInfo.CoverUrl) > 0 && oldCommunity.BasicInfo.CoverUrl != community.BasicInfo.CoverUrl {
		err = c.fileService.DeleteFile(ctx, path.Base(oldCommunity.BasicInfo.CoverUrl))
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("failed to delete community cover: %v", err))
			return nil, err
		}
	}

	return &updatedCommunity, nil
}

func (c *CommunityUseCase) JoinCommunity(ctx context.Context, member models.CommunityMember) error {
	if member.CommunityID == uuid.Nil || member.UserID == uuid.Nil {
		logger.Error(ctx, "user ID and community ID cannot be empty")
		return fmt.Errorf("user ID and community ID cannot be empty")
	}

	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now()
	}

	if len(member.Role) == 0 {
		member.Role = models.CommunityRoleMember
	}

	if err := c.repo.JoinCommunity(ctx, member); err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to join community: %v", err))
		return err
	}
	return nil
}

func (c *CommunityUseCase) LeaveCommunity(ctx context.Context, userId, communityId uuid.UUID) error {
	if userId == uuid.Nil || communityId == uuid.Nil {
		logger.Error(ctx, "user ID and community ID cannot be empty")
		return fmt.Errorf("user ID and community ID cannot be empty")
	}

	if err := c.repo.LeaveCommunity(ctx, userId, communityId); err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to leave community: %v", err))
		return err
	}
	return nil
}

func (c *CommunityUseCase) GetUserCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]models.Community, error) {
	if userId == uuid.Nil {
		logger.Error(ctx, "user ID cannot be empty")
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	communities, err := c.repo.GetUserCommunities(ctx, userId, count, ts)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get user communities: %v", err))
		return nil, err
	}

	return communities, nil
}

func (c *CommunityUseCase) SearchSimilarCommunities(ctx context.Context, name string, count int) ([]models.Community, error) {
	if len(name) == 0 {
		logger.Error(ctx, "community name cannot be empty")
		return nil, fmt.Errorf("community name cannot be empty")
	}

	if count <= 0 {
		logger.Error(ctx, "count must be greater than 0")
		return nil, fmt.Errorf("count must be greater than 0")
	}

	communities, err := c.repo.SearchSimilarCommunities(ctx, name, count)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to search similar communities: %v", err))
		return nil, err
	}

	return communities, nil
}

func (c *CommunityUseCase) ChangeUserRole(ctx context.Context, userId, communityId uuid.UUID, role models.CommunityRole, requester uuid.UUID) error {
	if userId == uuid.Nil || communityId == uuid.Nil || requester == uuid.Nil {
		logger.Error(ctx, "user ID and community ID cannot be empty")
		return fmt.Errorf("user ID and community ID cannot be empty")
	}

	// check if requester is admin or owner
	isMember, requesterRole, err := c.repo.IsCommunityMember(ctx, requester, communityId)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to check if user is a community member: %v", err))
		return err
	}
	if !isMember || requesterRole == nil {
		logger.Error(ctx, "requester is not a member of the community")
		return community_errors.ErrNotParticipant
	}
	if *requesterRole != models.CommunityRoleAdmin && *requesterRole != models.CommunityRoleOwner {
		logger.Error(ctx, "requester is not admin or owner")
		return community_errors.ErrForbidden
	}

	if err := c.repo.ChangeUserRole(ctx, userId, communityId, role); err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to change user role: %v", err))
		return err
	}
	return nil
}

func (c *CommunityUseCase) GetControlledCommunities(ctx context.Context, userId uuid.UUID, count int, ts time.Time) ([]models.Community, error) {
	if userId == uuid.Nil {
		logger.Error(ctx, "user ID cannot be empty")
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	communities, err := c.repo.GetControlledCommunities(ctx, userId, count, ts)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("failed to get user communities: %v", err))
		return nil, err
	}

	return communities, nil
}
