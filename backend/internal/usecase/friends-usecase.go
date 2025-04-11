package usecase

import (
	"context"
	"quickflow/internal/models"
)

type FriendsRepository interface {
	GetFriendsPublicInfo(ctx context.Context, userID string) ([]models.FriendInfo, error)
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

func (f *FriendsService) GetFriendsInfo(ctx context.Context, userID string) ([]models.FriendInfo, error) {
	friendsIds, err := f.friendsRepo.GetFriendsPublicInfo(ctx, userID)
	if err != nil {
		return []models.FriendInfo{}, err
	}

	return friendsIds, nil
}
