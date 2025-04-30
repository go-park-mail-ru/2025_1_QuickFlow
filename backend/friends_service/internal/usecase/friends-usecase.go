package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"quickflow/internal/models"
	"quickflow/pkg/logger"
)

type FriendsRepository interface {
	GetFriendsPublicInfo(ctx context.Context, userID string, amount int, startPos int) ([]models.FriendInfo, int, error)
	SendFriendRequest(ctx context.Context, senderID string, receiverID string) error
	AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error
	DeleteFriend(ctx context.Context, senderID string, receiverID string) error
	Unfollow(ctx context.Context, userID string, friendID string) error
	IsExistsFriendRequest(ctx context.Context, senderID string, receiverID string) (bool, error)
	GetUserRelation(ctx context.Context, user1 uuid.UUID, user2 uuid.UUID) (models.UserRelation, error)
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

func (f *FriendsService) GetFriendsInfo(ctx context.Context, userID string, limit string, offset string) ([]models.FriendInfo, int, error) {
	amount, err := strconv.Atoi(limit)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to parse count. Given value %s: %s", limit, err.Error()))
		return nil, 0, err
	}

	startPos, err := strconv.Atoi(offset)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("Unable to parse offset. Given value %s: %s", offset, err.Error()))
		return nil, 0, err
	}

	friendsIds, friendsCount, err := f.friendsRepo.GetFriendsPublicInfo(ctx, userID, amount, startPos)
	if err != nil {
		return []models.FriendInfo{}, 0, err
	}

	return friendsIds, friendsCount, err
}

func (f *FriendsService) SendFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	if err := f.friendsRepo.SendFriendRequest(ctx, senderID, receiverID); err != nil {
		return err
	}

	return nil
}

func (f *FriendsService) IsExistsFriendRequest(ctx context.Context, senderID string, receiverID string) (bool, error) {
	isExists, err := f.friendsRepo.IsExistsFriendRequest(ctx, senderID, receiverID)
	if err != nil {
		return false, err
	}

	return isExists, nil
}

func (f *FriendsService) AcceptFriendRequest(ctx context.Context, senderID string, receiverID string) error {
	if err := f.friendsRepo.AcceptFriendRequest(ctx, senderID, receiverID); err != nil {
		return err
	}

	return nil
}

func (f *FriendsService) DeleteFriend(ctx context.Context, userID string, friendID string) error {
	if err := f.friendsRepo.DeleteFriend(ctx, userID, friendID); err != nil {
		return err
	}

	return nil
}

func (f *FriendsService) GetUserRelation(ctx context.Context, user1 uuid.UUID, user2 uuid.UUID) (models.UserRelation, error) {
	if user1 == uuid.Nil || user2 == uuid.Nil {
		return models.RelationStranger, fmt.Errorf("userID is empty")
	}

	if user1 == user2 {
		return models.RelationSelf, nil
	}

	relation, err := f.friendsRepo.GetUserRelation(ctx, user1, user2)
	if err != nil {
		return models.RelationStranger, fmt.Errorf("f.friendsRepo.GetUserRelation: %w", err)
	}

	return relation, nil
}

func (f *FriendsService) Unfollow(ctx context.Context, userID string, friendID string) error {
	if err := f.friendsRepo.Unfollow(ctx, userID, friendID); err != nil {
		return err
	}

	return nil
}
