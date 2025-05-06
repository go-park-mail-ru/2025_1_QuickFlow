package postgres_models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestCommunityModelToPostgres(t *testing.T) {
	community := &models.Community{
		ID:        uuid.New(),
		NickName:  "TestCommunity",
		OwnerID:   uuid.New(),
		CreatedAt: time.Now(),
		BasicInfo: &models.BasicCommunityInfo{
			Name:        "Test Community",
			Description: "Description of community",
			AvatarUrl:   "http://avatar.url",
			CoverUrl:    "http://cover.url",
		},
	}

	communityPostgres := CommunityModelToPostgres(community)

	assert.True(t, community.ID == communityPostgres.Id.Bytes)
	assert.True(t, community.OwnerID == communityPostgres.OwnerId.Bytes)
	assert.Equal(t, community.NickName, communityPostgres.NickName.String)
	assert.Equal(t, community.BasicInfo.Name, communityPostgres.Name.String)
	assert.Equal(t, community.BasicInfo.Description, communityPostgres.Description.String)
	assert.Equal(t, community.BasicInfo.AvatarUrl, communityPostgres.AvatarUrl.String)
	assert.Equal(t, community.BasicInfo.CoverUrl, communityPostgres.CoverUrl.String)
	assert.True(t, communityPostgres.CreatedAt.Valid)
}

func TestPostgresToCommunityModel(t *testing.T) {
	communityPostgres := &CommunityPostgres{
		Id:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
		OwnerId:     pgtype.UUID{Bytes: uuid.New(), Valid: true},
		NickName:    pgtype.Text{String: "TestCommunity", Valid: true},
		Name:        pgtype.Text{String: "Test Community", Valid: true},
		Description: pgtype.Text{String: "Description of community", Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		AvatarUrl:   pgtype.Text{String: "http://avatar.url", Valid: true},
		CoverUrl:    pgtype.Text{String: "http://cover.url", Valid: true},
	}

	communityModel := communityPostgres.PostgresToCommunityModel()

	assert.True(t, communityPostgres.Id.Bytes == communityModel.ID)
	assert.True(t, communityPostgres.OwnerId.Bytes == communityModel.OwnerID)
	assert.Equal(t, communityPostgres.NickName.String, communityModel.NickName)
	assert.Equal(t, communityPostgres.Name.String, communityModel.BasicInfo.Name)
	assert.Equal(t, communityPostgres.Description.String, communityModel.BasicInfo.Description)
	assert.Equal(t, communityPostgres.AvatarUrl.String, communityModel.BasicInfo.AvatarUrl)
	assert.Equal(t, communityPostgres.CoverUrl.String, communityModel.BasicInfo.CoverUrl)
	assert.True(t, communityModel.CreatedAt.Equal(communityPostgres.CreatedAt.Time))
}

func TestCommunityMemberModelToPostgres(t *testing.T) {
	member := &models.CommunityMember{
		UserID:      uuid.New(),
		CommunityID: uuid.New(),
		Role:        models.CommunityRoleOwner,
		JoinedAt:    time.Now(),
	}

	memberPostgres := CommunityMemberModelToPostgres(member)

	assert.True(t, member.UserID == memberPostgres.UserId.Bytes)
	assert.True(t, member.CommunityID == memberPostgres.CommunityId.Bytes)
	assert.Equal(t, string(member.Role), memberPostgres.Role.String)
	assert.True(t, memberPostgres.JoinedAt.Valid)
}

func TestPostgresToCommunityMemberModel(t *testing.T) {
	memberPostgres := &CommunityMemberPostgres{
		UserId:      pgtype.UUID{Bytes: uuid.New(), Valid: true},
		CommunityId: pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Role:        pgtype.Text{String: "owner", Valid: true},
		JoinedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	memberModel := memberPostgres.PostgresToCommunityMemberModel()

	assert.True(t, memberPostgres.UserId.Bytes == memberModel.UserID)
	assert.True(t, memberPostgres.CommunityId.Bytes == memberModel.CommunityID)
	assert.Equal(t, models.CommunityRole(memberPostgres.Role.String), memberModel.Role)
	assert.True(t, memberModel.JoinedAt.Equal(memberPostgres.JoinedAt.Time))
}
