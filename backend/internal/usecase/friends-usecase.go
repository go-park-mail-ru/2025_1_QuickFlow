package usecase

import (
	"context"
	"github.com/google/uuid"
	"quickflow/internal/models"
)

type FriendsRepository interface {
	GetFriends(ctx context.Context, userId uuid.UUID) ([]string, error)
	GetFriendsInfo(ctx context.Context, friendIds []string) ([]models.FriendInfo, error)
}

type FriendsService struct {
	friendsRepo FriendsRepository
}

// NewFriendsService creates new friends service.
func NewFriendsService(friendsRepo FriendsRepository) *FriendsService {
	return &FriendsService{
		friendsRepo: friendsRepo,
	}
}

func (f *FriendsService) GetFriendsIds(ctx context.Context, userID uuid.UUID) ([]string, error) {
	friendsIds, err := f.friendsRepo.GetFriends(ctx, userID)
	if err != nil {
		return []string{}, err
	}

	return friendsIds, nil
}

func (f *FriendsService) GetFriendsInfo(ctx context.Context, friendIDs []string) ([]models.FriendInfo, error) {
	friendsIds, err := f.friendsRepo.GetFriendsInfo(ctx, friendIDs)
	if err != nil {
		return []models.FriendInfo{}, err
	}

	return friendsIds, nil
}
