package in_memory

import (
    "context"
    "errors"
    "slices"
    "time"

    "github.com/google/uuid"

    "quickflow/internal/models"
    tsslice "quickflow/pkg/thread-safe-slice"
)

type InMemoryPostRepository struct {
    posts *tsslice.ThreadSafeSlice[models.Post]
}

// NewInMemoryPostRepository creates new storage instance.
func NewInMemoryPostRepository() *InMemoryPostRepository {
    return &InMemoryPostRepository{
        posts: tsslice.NewThreadSafeSlice[models.Post](),
    }
}

// AddPost adds post to the repository.
func (r *InMemoryPostRepository) AddPost(ctx context.Context, post models.Post) error {
    r.posts.Add(post)
    return nil
}

// DeletePost removes post from the repository.
func (r *InMemoryPostRepository) DeletePost(ctx context.Context, postId uuid.UUID) error {
    return r.posts.DeleteIf(func(post models.Post) bool {
        return post.Id == postId
    })
}

// GetPostsForUId returns posts for user.
func (r *InMemoryPostRepository) GetPostsForUId(ctx context.Context, uid uuid.UUID, numPosts int, timestamp time.Time) ([]models.Post, error) {
    posts := r.posts.Filter(func(post models.Post) bool {
        return post.CreatedAt.Before(timestamp)
    }, numPosts)

    if len(posts) == 0 {
        return nil, errors.New("no posts found for user")
    }

    slices.SortFunc(posts, func(p1, p2 models.Post) int {
        if p1.CreatedAt.Before(p2.CreatedAt) {
            return 1
        }
        if p1.CreatedAt.After(p2.CreatedAt) {
            return -1
        }
        return 0
    })
    return posts, nil
}
