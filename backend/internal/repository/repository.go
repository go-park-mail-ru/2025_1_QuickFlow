package repository

//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"github.com/google/uuid"
//	"quickflow/internal/models"
//	"sync"
//	"time"
//)
//
//type InMemory struct {
//	users    map[uuid.UUID]models.User
//	sessions map[string]uuid.UUID
//	mu       sync.RWMutex
//	posts    []models.Post
//}
//
//// NewInMemory creates new storage instance
//func NewInMemory() *InMemory {
//	return &InMemory{
//		users:    make(map[uuid.UUID]models.User),
//		sessions: make(map[string]uuid.UUID),
//	}
//}
//
//// LookupUserSession returns user by session.
//func (i *InMemory) LookupUserSession(_ context.Context, session models.Session) (models.User, error) {
//	i.mu.RLock()
//	defer i.mu.RUnlock()
//
//	userId, found := i.sessions[session.SessionId]
//	if !found {
//		return models.User{}, errors.New("session not found")
//	}
//
//	user, found := i.users[userId]
//	if !found {
//		return models.User{}, errors.New("user not found")
//	}
//
//	return user, nil
//}
//
//// AddPost adds post to the repository.
//func (i *InMemory) AddPost(_ context.Context, post models.Post) error {
//	i.mu.Lock()
//	defer i.mu.Unlock()
//
//	i.posts = append(i.posts, post)
//	return nil
//}
//
//// DeletePost removes post from the repository.
//func (i *InMemory) DeletePost(_ context.Context, postId uuid.UUID) error {
//	i.mu.Lock()
//	defer i.mu.Unlock()
//
//	for idx, post := range i.posts {
//		if post.Id == postId {
//			i.posts = append(i.posts[:idx], i.posts[idx+1:]...)
//			return nil
//		}
//	}
//
//	return errors.New("post not found")
//}
//
//// GetPostsForUId returns all posts to be shown to the user.
//func (i *InMemory) GetPostsForUId(_ context.Context, uid int64, numPosts int, timestamp time.Time) ([]models.Post, error) {
//	i.mu.RLock()
//	defer i.mu.RUnlock()
//
//	var result []models.Post
//	for _, post := range i.posts {
//		if len(result) == numPosts {
//			break
//		}
//
//		// TODO: Пока выводим все посты, что есть
//		if post.CreatedAt.Before(timestamp) {
//			result = append(result, post)
//		}
//	}
//
//	if len(result) == 0 {
//		return nil, errors.New("no posts found for user")
//	}
//
//	return result, nil
//}
//
//func (i *InMemory) SaveUser(user models.User) (uuid.UUID, models.Session, error) {
//	session := models.CreateSession(i.sessions)
//	user = models.CreateUser(user, i.users)
//
//	i.sessions[session.SessionId] = user.Id
//	i.users[user.Login] = user
//
//	return user.Id, session, nil
//}
//
//func (i *InMemory) GetUser(authData models.AuthForm) (models.Session, error) {
//	user, exists := i.users[authData.Login]
//
//	if !exists || !models.CheckPassword(authData.Password, user) {
//		return models.Session{}, fmt.Errorf("incorrect login or password")
//	}
//
//	newSession := models.CreateSession(i.sessions)
//	i.sessions[newSession.SessionId] = user.Id
//
//	return newSession, nil
//
//}
