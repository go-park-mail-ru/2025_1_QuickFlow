package in_memory

import (
    "context"
    "errors"
    "sync"
    "time"

    "github.com/google/uuid"

    "quickflow/internal/models"
)

type InMemoryPostRepository struct {
    mu    sync.RWMutex
    posts []models.Post
}

func NewInMemoryPostRepository() *InMemoryPostRepository {
    return &InMemoryPostRepository{
        posts: make([]models.Post, 0),
    }
}

func (r *InMemoryPostRepository) AddPost(ctx context.Context, post models.Post) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.posts = append(r.posts, post)
    return nil
}

func (r *InMemoryPostRepository) DeletePost(ctx context.Context, postId uuid.UUID) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    for idx, post := range r.posts {
        if post.Id == postId {
            r.posts = append(r.posts[:idx], r.posts[idx+1:]...)
            return nil
        }
    }

    return errors.New("post not found")
}

func (r *InMemoryPostRepository) GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var result []models.Post
    for _, post := range r.posts {
        if len(result) == numPosts {
            break
        }

        // TODO: Пока выводим все посты, что есть
        if post.CreatedAt.Before(timestamp) {
            result = append(result, post)
        }
    }

    if len(result) == 0 {
        return nil, errors.New("no posts found for user")
    }

    return result, nil
}
