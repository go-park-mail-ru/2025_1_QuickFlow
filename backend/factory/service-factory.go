package factory

import (
	"quickflow/internal/usecase"
)

type DefaultServiceFactory struct {
	repoFactory RepositoryFactory
}

func NewDefaultServiceFactory(repoFactory RepositoryFactory) *DefaultServiceFactory {
	return &DefaultServiceFactory{
		repoFactory: repoFactory,
	}
}

func (f *DefaultServiceFactory) FeedBackService() *usecase.FeedbackService {
	return usecase.NewFeedBackService(
		f.repoFactory.FeedbackRepository(),
	)
}

func (f *DefaultServiceFactory) AuthService() *usecase.AuthService {
	return usecase.NewAuthService(
		f.repoFactory.UserRepository(),
		f.repoFactory.SessionRepository(),
		f.repoFactory.ProfileRepository(),
	)
}

func (f *DefaultServiceFactory) ProfileService() *usecase.ProfileService {
	return usecase.NewProfileService(
		f.repoFactory.ProfileRepository(),
		f.repoFactory.UserRepository(),
		f.repoFactory.FileRepository(),
	)
}

func (f *DefaultServiceFactory) PostService() *usecase.PostService {
	return usecase.NewPostService(
		f.repoFactory.PostRepository(),
		f.repoFactory.FileRepository(),
	)
}

func (f *DefaultServiceFactory) ChatService() *usecase.ChatService {
	return usecase.NewChatUseCase(
		f.repoFactory.ChatRepository(),
		f.repoFactory.FileRepository(),
		f.repoFactory.ProfileRepository(),
		f.repoFactory.MessageRepository(),
	)
}

func (f *DefaultServiceFactory) MessageService() *usecase.MessageUseCase {
	return usecase.NewMessageUseCase(
		f.repoFactory.MessageRepository(),
		f.repoFactory.FileRepository(),
		f.repoFactory.ChatRepository(),
	)
}

func (f *DefaultServiceFactory) FriendService() *usecase.FriendsService {
	return usecase.NewFriendsService(
		f.repoFactory.FriendRepository(),
	)
}

func (f *DefaultServiceFactory) SearchService() *usecase.SearchService {
	return usecase.NewSearchService(
		f.repoFactory.UserRepository(),
	)
}
