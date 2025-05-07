package forms

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestCreateFormToModel(t *testing.T) {
	tests := []struct {
		name     string
		form     CreateCommunityForm
		expected models.Community
	}{
		{
			name: "valid form",
			form: CreateCommunityForm{
				Nickname:    "TestCommunity",
				Name:        "Test Community",
				Description: "A test community",
			},
			expected: models.Community{
				NickName: "TestCommunity",
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Test Community",
					Description: "A test community",
				},
			},
		},
		{
			name: "form with nil avatar and cover",
			form: CreateCommunityForm{
				Nickname:    "AnotherCommunity",
				Name:        "Another Community",
				Description: "Another test community",
				Avatar:      nil,
				Cover:       nil,
			},
			expected: models.Community{
				NickName: "AnotherCommunity",
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Another Community",
					Description: "Another test community",
				},
				Avatar: nil,
				Cover:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.form.CreateFormToModel()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestToCommunityForm(t *testing.T) {
	tests := []struct {
		name      string
		community models.Community
		ownerInfo models.PublicUserInfo
		expected  CommunityForm
	}{
		{
			name: "valid community form",
			community: models.Community{
				ID:       uuid.New(),
				NickName: "TestCommunity",
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Test Community",
					Description: "A test community",
				},
				CreatedAt: time.Now(),
			},
			ownerInfo: models.PublicUserInfo{
				Id:        uuid.New(),
				Username:  "testowner",
				Firstname: "Test",
				Lastname:  "Owner",
				AvatarURL: "avatar_url",
			},
			expected: CommunityForm{
				CommunityInfo: &CommunityInfo{
					NickName:    "TestCommunity",
					Name:        "Test Community",
					Description: "A test community",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ToCommunityForm(tt.community, tt.ownerInfo)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestPaginationForm_GetParams(t *testing.T) {
	tests := []struct {
		name        string
		values      url.Values
		expected    PaginationForm
		expectError bool
	}{
		{
			name: "valid params",
			values: url.Values{
				"count": {"10"},
				"ts":    {time.Now().Format(time.RFC3339)},
			},
			expected: PaginationForm{
				Count: 10,
				Ts:    time.Now(),
			},
			expectError: false,
		},
		{
			name: "missing count param",
			values: url.Values{
				"ts": {time.Now().Format(time.RFC3339)},
			},
			expected:    PaginationForm{},
			expectError: true,
		},
		{
			name: "invalid count param",
			values: url.Values{
				"count": {"invalid"},
				"ts":    {time.Now().Format(time.RFC3339)},
			},
			expected:    PaginationForm{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form PaginationForm
			err := form.GetParams(tt.values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, form)
			}
		})
	}
}

func TestCommunityInfoFromModel(t *testing.T) {
	communityInfo := models.BasicCommunityInfo{
		Name:        "Test Community",
		Description: "A test community",
		AvatarUrl:   "avatar_url",
		CoverUrl:    "cover_url",
	}

	nickName := "TestCommunity"
	expected := &CommunityInfo{
		NickName:    nickName,
		Name:        communityInfo.Name,
		Description: communityInfo.Description,
		AvatarUrl:   communityInfo.AvatarUrl,
		CoverUrl:    communityInfo.CoverUrl,
	}

	actual := CommunityInfoFromModel(communityInfo, nickName)

	assert.Equal(t, expected, actual)
}

func TestToCommunityMemberOut(t *testing.T) {
	member := models.CommunityMember{
		UserID:      uuid.New(),
		CommunityID: uuid.New(),
		Role:        models.CommunityRoleMember,
		JoinedAt:    time.Now(),
	}

	info := models.PublicUserInfo{
		Id:        uuid.New(),
		Username:  "user123",
		Firstname: "John",
		Lastname:  "Doe",
		AvatarURL: "avatar_url",
	}

	expected := CommunityMemberOut{
		UserId:      member.UserID.String(),
		CommunityId: member.CommunityID.String(),
		Role:        string(member.Role),
		JoinedAt:    member.JoinedAt.Format(time.RFC3339),
		PublicUserInfoOut: PublicUserInfoOut{
			ID:        info.Id.String(),
			Username:  info.Username,
			FirstName: info.Firstname,
			LastName:  info.Lastname,
			AvatarURL: info.AvatarURL,
		},
	}

	actual := ToCommunityMemberOut(member, info)

	assert.Equal(t, expected, actual)
}
