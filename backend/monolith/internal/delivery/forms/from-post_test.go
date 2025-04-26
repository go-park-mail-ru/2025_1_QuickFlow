package forms

import (
	"quickflow/monolith/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPostOut_FromPost(t *testing.T) {
	tests := []struct {
		name     string
		post     models.Post
		expected PostOut
	}{
		{
			name: "Valid Post transformation",
			post: models.Post{
				Id:           uuid.New(),
				CreatorId:    uuid.New(),
				Desc:         "Test Post Description",
				ImagesURL:    []string{"pic1.jpg", "pic2.jpg"},
				CreatedAt:    time.Date(2025, 3, 11, 15, 30, 0, 0, time.UTC),
				LikeCount:    10,
				RepostCount:  5,
				CommentCount: 3,
			},
			expected: PostOut{
				Desc:         "Test Post Description",
				Pics:         []string{"pic1.jpg", "pic2.jpg"},
				CreatedAt:    "2025-03-11T15:30:00Z", // Сформатированная дата
				LikeCount:    10,
				RepostCount:  5,
				CommentCount: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var postOut PostOut
			// Преобразуем post в postOut
			postOut.FromPost(tt.post)

			// Проверка Id и CreatorId (UUID)
			if postOut.Id != tt.post.Id.String() {
				t.Errorf("FromPost() Id = %v, want %v", postOut.Id, tt.post.Id.String())
			}
			if postOut.Creator.ID != tt.post.CreatorId.String() {
				t.Errorf("FromPost() CreatorId = %v, want %v", postOut.Creator.ID, tt.post.CreatorId.String())
			}

			// Проверка остальных полей
			if postOut.Desc != tt.expected.Desc {
				t.Errorf("FromPost() Text = %v, want %v", postOut.Desc, tt.expected.Desc)
			}
			if !equalStringSlices(postOut.Pics, tt.expected.Pics) {
				t.Errorf("FromPost() Files = %v, want %v", postOut.Pics, tt.expected.Pics)
			}
			if postOut.CreatedAt != tt.expected.CreatedAt {
				t.Errorf("FromPost() CreatedAt = %v, want %v", postOut.CreatedAt, tt.expected.CreatedAt)
			}
			if postOut.LikeCount != tt.expected.LikeCount {
				t.Errorf("FromPost() LikeCount = %v, want %v", postOut.LikeCount, tt.expected.LikeCount)
			}
			if postOut.RepostCount != tt.expected.RepostCount {
				t.Errorf("FromPost() RepostCount = %v, want %v", postOut.RepostCount, tt.expected.RepostCount)
			}
			if postOut.CommentCount != tt.expected.CommentCount {
				t.Errorf("FromPost() CommentCount = %v, want %v", postOut.CommentCount, tt.expected.CommentCount)
			}
		})
	}
}

// Вспомогательная функция для сравнения слайсов строк
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
