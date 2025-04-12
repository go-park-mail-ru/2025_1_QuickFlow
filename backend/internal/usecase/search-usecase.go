package usecase

import (
	"context"

	"quickflow/internal/models"
)

type SearchService struct {
	userRepo UserRepository
}

func NewSearchService(userRepo UserRepository) *SearchService {
	return &SearchService{
		userRepo: userRepo,
	}
}

func (s *SearchService) SearchSimilarUser(ctx context.Context, username string, postsCount uint) ([]models.PublicUserInfo, error) {
	users, err := s.userRepo.SearchSimilar(ctx, username, postsCount)
	if err != nil {
		return nil, err
	}

	return users, nil
}
