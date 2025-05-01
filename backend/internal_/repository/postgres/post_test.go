package postgres_test

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"quickflow/internal/models"
	"quickflow/internal/repository/postgres"
	postgresmodels "quickflow/internal/repository/postgres/postgres-models"
	"testing"
	"time"
)

func TestPostRepository(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		post      models.Post
		mockSetup func(mock sqlmock.Sqlmock, post models.Post)
		wantErr   bool
	}{
		{
			name: "success add post",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				pgPost := postgresmodels.ConvertPostToPostgres(post)
				mock.ExpectExec(`(?i)INSERT INTO post`).
					WithArgs(
						pgPost.Id,
						pgPost.CreatorId,
						pgPost.Desc,
						pgPost.CreatedAt,
						pgPost.UpdatedAt,
						pgPost.LikeCount,
						pgPost.RepostCount,
						pgPost.CommentCount,
						pgPost.IsRepost,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				for _, pic := range post.ImagesURL {
					mock.ExpectExec(`(?i)INSERT INTO post_file`).
						WithArgs(pgPost.Id, pic).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
			},
			wantErr: false,
		},
		{
			name: "db error on add post",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				pgPost := postgresmodels.ConvertPostToPostgres(post)
				mock.ExpectExec(`(?i)INSERT INTO post`).
					WithArgs(pgPost.Id, pgPost.CreatorId, pgPost.Desc, pgPost.CreatedAt, pgPost.UpdatedAt, pgPost.LikeCount, pgPost.RepostCount, pgPost.CommentCount, pgPost.IsRepost).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "success delete post",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				mock.ExpectExec(`(?i)DELETE FROM post`).
					WithArgs(post.Id).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "db error on delete post",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				mock.ExpectExec(`(?i)DELETE FROM post`).
					WithArgs(post.Id).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "success get post",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				pgPost := postgresmodels.ConvertPostToPostgres(post)
				mock.ExpectQuery(`(?i)select p.id, creator_id, text, created_at, updated_at, like_count, repost_count, comment_count, is_repost`).
					WithArgs(pgPost.Id).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "creator_id", "text", "created_at", "updated_at", "like_count", "repost_count", "comment_count", "is_repost",
					}).AddRow(pgPost.Id, pgPost.CreatorId, pgPost.Desc, pgPost.CreatedAt, pgPost.UpdatedAt, pgPost.LikeCount, pgPost.RepostCount, pgPost.CommentCount, pgPost.IsRepost))

				mock.ExpectQuery(`(?i)SELECT file_url`).
					WithArgs(pgPost.Id).
					WillReturnRows(sqlmock.NewRows([]string{"file_url"}).AddRow(post.ImagesURL[0]))
			},
			wantErr: false,
		},
		{
			name: "db error on get post",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				mock.ExpectQuery(`(?i)SELECT id, creator_id, Desc, created_at, updated_at, like_count, repost_count, comment_count, is_repost`).
					WithArgs(post.Id).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "success update post Desc",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				mock.ExpectExec(`(?i)UPDATE post`).
					WithArgs(post.Desc, sqlmock.AnyArg(), post.Id).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "db error on update post Desc",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				mock.ExpectExec(`(?i)UPDATE post`).
					WithArgs(post.Desc, sqlmock.AnyArg(), post.Id).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "success update post files",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				mock.ExpectExec(`(?i)DELETE FROM post_file`).
					WithArgs(post.Id).
					WillReturnResult(sqlmock.NewResult(1, 1))

				for _, fileURL := range post.ImagesURL {
					mock.ExpectExec(`(?i)INSERT INTO post_file`).
						WithArgs(post.Id, fileURL).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
			},
			wantErr: false,
		},
		{
			name: "db error on update post files",
			post: newTestPost(),
			mockSetup: func(mock sqlmock.Sqlmock, post models.Post) {
				mock.ExpectExec(`(?i)DELETE FROM post_file`).
					WithArgs(post.Id).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			require.NoError(t, err)

			repo := postgres.NewPostgresPostRepository(mockDB)
			tt.mockSetup(mock, tt.post)

			switch tt.name {
			case "success add post":
				err = repo.AddPost(ctx, tt.post)
			case "db error on add post":
				err = repo.AddPost(ctx, tt.post)
			case "success delete post":
				err = repo.DeletePost(ctx, tt.post.Id)
			case "db error on delete post":
				err = repo.DeletePost(ctx, tt.post.Id)
			case "success get post":
				_, err = repo.GetPost(ctx, tt.post.Id)
			case "db error on get post":
				_, err = repo.GetPost(ctx, tt.post.Id)
			case "success update post Desc":
				err = repo.UpdatePostText(ctx, tt.post.Id, tt.post.Desc)
			case "db error on update post Desc":
				err = repo.UpdatePostText(ctx, tt.post.Id, tt.post.Desc)
			case "success update post files":
				err = repo.UpdatePostFiles(ctx, tt.post.Id, tt.post.ImagesURL)
			case "db error on update post files":
				err = repo.UpdatePostFiles(ctx, tt.post.Id, tt.post.ImagesURL)
			}

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func newTestPost() models.Post {
	return models.Post{
		Id:           uuid.New(),
		CreatorId:    uuid.New(),
		Desc:         "Test Post",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LikeCount:    10,
		RepostCount:  5,
		CommentCount: 2,
		IsRepost:     false,
		ImagesURL:    []string{"http://example.com/image1.jpg"},
	}
}
