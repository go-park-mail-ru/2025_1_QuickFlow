package in_memory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"quickflow/internal/models"
)

func TestInMemoryPostRepository_AddPost(t *testing.T) {
	repo := NewInMemoryPostRepository()

	tests := []struct {
		name    string
		post    models.Post
		wantErr bool
	}{
		{
			name:    "Add valid post",
			post:    models.Post{Id: uuid.New(), CreatorId: uuid.New(), Desc: "Hello", CreatedAt: time.Now()},
			wantErr: false,
		},
		{
			name:    "Add empty post",
			post:    models.Post{Id: uuid.Nil, CreatorId: uuid.Nil, Desc: "", CreatedAt: time.Now()},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.AddPost(context.Background(), tt.post)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddPost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInMemoryPostRepository_DeletePost(t *testing.T) {
	repo := NewInMemoryPostRepository()
	postId := uuid.New()
	repo.AddPost(context.Background(), models.Post{Id: postId, CreatorId: uuid.New(), Desc: "Test Post", CreatedAt: time.Now()})

	tests := []struct {
		name    string
		postId  uuid.UUID
		wantErr bool
	}{
		{
			name:    "Delete existing post",
			postId:  postId,
			wantErr: false,
		},
		{
			name:    "Delete non-existing post",
			postId:  uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.DeletePost(context.Background(), tt.postId)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeletePost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInMemoryPostRepository_GetPostsForUId(t *testing.T) {
	repo := NewInMemoryPostRepository()

	CreatorId := uuid.New()
	repo.AddPost(context.Background(), models.Post{Id: uuid.New(), CreatorId: CreatorId, Desc: "Post 1", CreatedAt: time.Now().Add(-1 * time.Hour)})
	repo.AddPost(context.Background(), models.Post{Id: uuid.New(), CreatorId: CreatorId, Desc: "Post 2", CreatedAt: time.Now()})

	tests := []struct {
		name      string
		uid       uuid.UUID
		numPosts  int
		timestamp time.Time
		wantErr   bool
	}{
		{
			name:      "Get posts for existing user",
			uid:       CreatorId,
			numPosts:  2,
			timestamp: time.Now(),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.GetPostsForUId(context.Background(), tt.uid, tt.numPosts, tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPostsForUId() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
