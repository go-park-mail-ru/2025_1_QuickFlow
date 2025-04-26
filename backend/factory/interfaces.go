package factory

import (
	"quickflow/internal/usecase"
)

type RepositoryFactory interface {
	UserRepository() usecase.UserRepository
	SessionRepository() usecase.SessionRepository
	ProfileRepository() usecase.ProfileRepository
	PostRepository() usecase.PostRepository
	ChatRepository() usecase.ChatRepository
	MessageRepository() usecase.MessageRepository
	FileRepository() usecase.FileRepository
	FriendRepository() usecase.FriendsRepository
	FeedbackRepository() usecase.FeedbackRepository
	Close() error
}

type ServiceFactory interface {
	AuthService() *usecase.AuthService
	ProfileService() *usecase.ProfileService
	PostService() *usecase.PostService
	ChatService() *usecase.ChatService
	MessageService() *usecase.MessageService
	FriendService() *usecase.FriendsService
	SearchService() *usecase.SearchService
	FeedBackService() *usecase.FeedbackService
}

type HandlerFactory interface {
	InitHttpHandlers() *HttpHandlerCollection
	InitWSHandlers() *WSHandlerCollection
}
