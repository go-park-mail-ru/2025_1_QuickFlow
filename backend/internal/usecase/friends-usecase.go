package usecase

import (
	"context"
	"fmt"
	"quickflow/internal/models"
	"quickflow/pkg/logger"
	"strconv"
)

type FriendsRepository interface {
	GetFriendsPublicInfo(ctx context.Context, userID string, amount int, startPos int) ([]models.FriendInfo, bool, error)
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

func (f *FriendsService) GetFriendsInfo(ctx context.Context, userID string, limit string, offset string) ([]models.FriendInfo, bool, error) {
	amount, err := strconv.Atoi(limit)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to parse count. Given value %s: %s", limit, err.Error()))
		return nil, false, err
	}

	startPos, err := strconv.Atoi(offset)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to parse offset. Given value %s: %s", offset, err.Error()))
		return nil, false, err
	}

	friendsIds, hasMore, err := f.friendsRepo.GetFriendsPublicInfo(ctx, userID, amount, startPos)
	if err != nil {
		return []models.FriendInfo{}, false, err
	}

	return friendsIds, hasMore, nil
}
