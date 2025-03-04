package usecase

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"

    "quickflow/internal/models"
)

type UserRepository interface {
    SaveUser(user models.User) (uuid.UUID, error)
    GetUser(authData models.AuthForm) (models.User, error)
    GetUserByUId(ctx context.Context, uid uuid.UUID) (models.User, error)
}

type PostRepository interface {
    AddPost(ctx context.Context, post models.Post) error
    DeletePost(ctx context.Context, postId uuid.UUID) error
    GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error)
}

type SessionRepository interface {
    CreateSession(userId uuid.UUID) models.Session
    LookupUserSession(ctx context.Context, session models.Session) (uuid.UUID, error)
}

type Processor struct {
    userRepo    UserRepository
    postRepo    PostRepository
    sessionRepo SessionRepository
}

// NewProcessor creates new Processor instance.
func NewProcessor(userRepo UserRepository, postRepo PostRepository, sessionRepo SessionRepository) *Processor {
    return &Processor{
        userRepo:    userRepo,
        postRepo:    postRepo,
        sessionRepo: sessionRepo,
    }
}

// LookupUserSession returns user by session.
func (p *Processor) LookupUserSession(ctx context.Context, session models.Session) (models.User, error) {
    userID, err := p.sessionRepo.LookupUserSession(ctx, session)
    if err != nil {
        return models.User{}, fmt.Errorf("p.repo.LookupUserSession: %w", err)
    }

    user, err := p.userRepo.GetUserByUId(ctx, userID)
    if err != nil {
        return models.User{}, fmt.Errorf("p.repo.GetUserByUId: %w", err)
    }

    return user, nil
}

// AddPost adds post to the repository.
func (p *Processor) AddPost(ctx context.Context, post models.Post) error {
    err := p.postRepo.AddPost(ctx, post)
    if err != nil {
        return fmt.Errorf("p.repo.AddPost: %w", err)
    }
    return nil
}

// DeletePost removes post from the repository.
func (p *Processor) DeletePost(ctx context.Context, postId uuid.UUID) error {
    err := p.postRepo.DeletePost(ctx, postId)
    if err != nil {
        return fmt.Errorf("p.repo.AddPost: %w", err)
    }
    return nil
}

// FetchFeed returns feed for user.
func (p *Processor) FetchFeed(ctx context.Context, user models.User, numPosts int, timestamp time.Time) ([]models.Post, error) {
    posts, err := p.postRepo.GetPostsForUId(ctx, user.Id, numPosts, timestamp)
    if err != nil {
        return []models.Post{}, fmt.Errorf("p.repo.GetPostsForUId: %w", err)
    }

    return posts, nil
}

func (p *Processor) CreateUser(user models.User) (uuid.UUID, models.Session, error) {
    userId, err := p.userRepo.SaveUser(user)
    if err != nil {
        return uuid.Nil, models.Session{}, err
    }

    session := p.sessionRepo.CreateSession(userId)

    return userId, session, nil
}

func (p *Processor) GetUser(authData models.AuthForm) (models.Session, error) {
    user, err := p.userRepo.GetUser(authData)
    if err != nil {
        return models.Session{}, err
    }

    session := p.sessionRepo.CreateSession(user.Id)

    return session, nil
}
