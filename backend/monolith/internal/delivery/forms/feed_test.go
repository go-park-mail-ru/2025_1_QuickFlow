package forms_test

import (
	"errors"
	"net/url"
	"quickflow/monolith/internal/delivery/forms"
	models2 "quickflow/monolith/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPostForm_ToPostModel(t *testing.T) {
	userID := uuid.New()
	postForm := forms.PostForm{
		Text:     "Hello, world!",
		IsRepost: true,
	}

	// Simulate the form having images (empty in this case)
	postForm.Images = []*models2.File{
		{Name: "image1.jpg", Reader: strings.NewReader("hi")},
	}

	post := postForm.ToPostModel(userID)

	assert.Equal(t, post.Desc, "Hello, world!")
	assert.Equal(t, post.CreatorId, userID)
	assert.Equal(t, post.IsRepost, true)
	assert.NotNil(t, post.Images)
	assert.Len(t, post.Images, 1)
}

func TestFeedForm_GetParams(t *testing.T) {
	tests := []struct {
		name          string
		values        url.Values
		expectedForm  forms.FeedForm
		expectedError error
	}{
		{
			name: "success with valid parameters",
			values: url.Values{
				"posts_count": []string{"5"},
				"ts":          []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm: forms.FeedForm{
				Posts: 5,
				Ts:    "2025-04-16T00:00:00Z",
			},
			expectedError: nil,
		},
		{
			name: "missing posts_count parameter",
			values: url.Values{
				"ts": []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm:  forms.FeedForm{},
			expectedError: errors.New("posts_count parameter missing"),
		},
		{
			name: "invalid posts_count format",
			values: url.Values{
				"posts_count": []string{"invalid"},
				"ts":          []string{"2025-04-16T00:00:00Z"},
			},
			expectedForm:  forms.FeedForm{},
			expectedError: errors.New("failed to parse posts_count"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form forms.FeedForm
			err := form.GetParams(tt.values)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedForm, form)
			}
		})
	}
}

func TestPublicUserInfoToOut(t *testing.T) {
	userID := uuid.New()
	// Create a mock PublicUserInfo
	publicUserInfo := models2.PublicUserInfo{
		Id:        userID,
		Username:  "user1",
		Firstname: "John",
		Lastname:  "Doe",
		AvatarURL: "http://example.com/avatar.jpg",
	}

	relation := models2.RelationFriend

	publicUserOut := forms.PublicUserInfoToOut(publicUserInfo, relation)

	assert.Equal(t, publicUserOut.ID, userID.String())
	assert.Equal(t, publicUserOut.Username, "user1")
	assert.Equal(t, publicUserOut.FirstName, "John")
	assert.Equal(t, publicUserOut.LastName, "Doe")
	assert.Equal(t, publicUserOut.AvatarURL, "http://example.com/avatar.jpg")
	assert.Equal(t, publicUserOut.Relation, models2.RelationFriend)
}

func TestPostOut_FromPost(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()

	now := time.Now()
	post := models2.Post{
		Id:        postID,
		CreatorId: userID,
		Desc:      "Sample post",
		ImagesURL: []string{"http://example.com/image.jpg"},
		CreatedAt: now,
		UpdatedAt: now,
		LikeCount: 10,
		IsRepost:  false,
	}

	postOut := forms.PostOut{}
	postOut.FromPost(post)

	assert.Equal(t, postOut.Id, postID.String())
	assert.Equal(t, postOut.Creator.ID, userID.String())
	assert.Equal(t, postOut.Desc, "Sample post")
	assert.Len(t, postOut.Pics, 1)
	assert.Equal(t, postOut.Pics[0], "http://example.com/image.jpg")
	assert.Equal(t, postOut.LikeCount, 10)
	assert.Equal(t, postOut.IsRepost, false)
}

func TestUpdatePostForm_ToPostUpdateModel(t *testing.T) {
	postID := uuid.New()
	updateForm := forms.UpdatePostForm{
		Text: "Updated post",
		Images: []*models2.File{
			{Name: "image1.jpg", Reader: strings.NewReader("hi")},
		},
	}

	postUpdate, err := updateForm.ToPostUpdateModel(postID)

	assert.NoError(t, err)
	assert.Equal(t, postUpdate.Id, postID)
	assert.Equal(t, postUpdate.Desc, "Updated post")
	assert.Len(t, postUpdate.Files, 1)
	assert.Equal(t, postUpdate.Files[0].Name, "image1.jpg")
}
