package usecase_test

import (
	"context"
	"errors"
	"quickflow/monolith/internal/models"
	"quickflow/monolith/internal/usecase"
	"quickflow/monolith/internal/usecase/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPostService_AddPost(t *testing.T) {
	tests := []struct {
		name           string
		post           models.Post
		uploadFilesErr error
		addPostErr     error
		expectedPost   models.Post
		expectedErr    error
	}{
		{
			name: "success",
			post: models.Post{
				Desc: "Hi",
			},
			expectedPost: models.Post{
				Desc: "Hi",
			},
			expectedErr: nil,
		},
		{
			name:        "add post error",
			post:        models.Post{Images: []*models.File{}},
			addPostErr:  errors.New("add post error"),
			expectedErr: errors.New("p.postRepo.AddPost: add post error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPostRepo := mocks.NewMockPostRepository(ctrl)
			mockFileRepo := mocks.NewMockFileRepository(ctrl)

			if tt.uploadFilesErr != nil {
				mockFileRepo.EXPECT().UploadManyFiles(gomock.Any(), tt.post.Images).Return(nil, tt.uploadFilesErr)
			} else {
				mockFileRepo.EXPECT().UploadManyFiles(gomock.Any(), tt.post.Images).Return(nil, nil)
			}

			if tt.addPostErr != nil {
				mockPostRepo.EXPECT().AddPost(gomock.Any(), gomock.Any()).Return(tt.addPostErr)
			} else {
				mockPostRepo.EXPECT().AddPost(gomock.Any(), gomock.Any()).Return(nil)
			}

			postService := usecase.NewPostService(mockPostRepo, mockFileRepo)

			result, err := postService.AddPost(context.Background(), tt.post)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				// Задаем UUID ожидаемому посту из результата
				tt.expectedPost.Id = result.Id
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPost, result)
			}
		})
	}
}

func TestPostService_DeletePost(t *testing.T) {
	tests := []struct {
		name   string
		user   models.User
		postId uuid.UUID
		belongsTo       bool
		deletePostErr   error
		getPostFilesErr error
		deleteFileErr   error
		expectedErr     error
	}{
		{
			name:        "success",
			user:        models.User{Id: uuid.New(), Username: "testuser"},
			postId:      uuid.New(),
			belongsTo:   true,
			expectedErr: nil,
		},
		{
			name:        "post does not belong to user",
			user:        models.User{Id: uuid.New(), Username: "testuser"},
			postId:      uuid.New(),
			belongsTo:   false,
			expectedErr: usecase.ErrPostDoesNotBelongToUser,
		},
		{
			name:          "delete file error",
			user:          models.User{Id: uuid.New(), Username: "testuser"},
			postId:        uuid.New(),
			belongsTo:     true,
			deleteFileErr: errors.New("delete file error"),
			expectedErr:   errors.New("p.fileRepo.DeleteFile: delete file error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPostRepo := mocks.NewMockPostRepository(ctrl)
			mockFileRepo := mocks.NewMockFileRepository(ctrl)

			// В ожиданиях проверяем, что BelongsTo вызывается для пользователя и поста
			mockPostRepo.EXPECT().BelongsTo(gomock.Any(), tt.user.Id, tt.postId).Return(tt.belongsTo, nil)

			if tt.belongsTo { // только если пост принадлежит пользователю
				if tt.getPostFilesErr != nil {
					mockPostRepo.EXPECT().GetPostFiles(gomock.Any(), tt.postId).Return(nil, tt.getPostFilesErr)
				} else {
					mockPostRepo.EXPECT().GetPostFiles(gomock.Any(), tt.postId).Return([]string{"file1.jpg"}, nil)
				}

				if tt.deletePostErr != nil {
					mockPostRepo.EXPECT().DeletePost(gomock.Any(), tt.postId).Return(tt.deletePostErr)
				} else {
					mockPostRepo.EXPECT().DeletePost(gomock.Any(), tt.postId).Return(nil)
				}

				if tt.deleteFileErr != nil {
					mockFileRepo.EXPECT().DeleteFile(gomock.Any(), gomock.Any()).Return(tt.deleteFileErr)
				} else {
					mockFileRepo.EXPECT().DeleteFile(gomock.Any(), gomock.Any()).Return(nil)
				}
			}

			// Создаем сервис
			postService := usecase.NewPostService(mockPostRepo, mockFileRepo)

			// Вызов метода DeletePost
			err := postService.DeletePost(context.Background(), tt.user, tt.postId)

			// Сравниваем ожидаемую ошибку с фактической
			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
