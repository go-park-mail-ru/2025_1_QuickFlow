package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"quickflow/shared/models"
)

func TestCreateCommunity(t *testing.T) {
	tests := []struct {
		name      string
		community models.Community
		mock      func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "Successfully create community",
			community: models.Community{
				ID:        uuid.New(),
				OwnerID:   uuid.New(),
				NickName:  "TestCommunity",
				CreatedAt: time.Now(),
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Test Community",
					Description: "A description of the community",
					AvatarUrl:   "http://avatar.url",
					CoverUrl:    "http://cover.url",
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`insert into community`).WithArgs(
					sqlmock.AnyArg(), sqlmock.AnyArg(), "Test Community", "A description of the community",
					sqlmock.AnyArg(), "http://avatar.url", "http://cover.url", "TestCommunity").
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(`insert into community_user`).WithArgs(
					sqlmock.AnyArg(), sqlmock.AnyArg(), "owner", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Failed to create community",
			community: models.Community{
				ID:        uuid.New(),
				OwnerID:   uuid.New(),
				NickName:  "TestCommunity",
				CreatedAt: time.Now(),
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Test Community",
					Description: "A description of the community",
					AvatarUrl:   "http://avatar.url",
					CoverUrl:    "http://cover.url",
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`insert into community`).WithArgs(
					sqlmock.AnyArg(), sqlmock.AnyArg(), "Test Community", "A description of the community",
					sqlmock.AnyArg(), "http://avatar.url", "http://cover.url", "TestCommunity").
					WillReturnError(fmt.Errorf("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			repo := &SqlCommunityRepository{connPool: mockDB}
			tt.mock(mock)

			err = repo.CreateCommunity(context.Background(), tt.community)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCommunity() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetCommunityById(t *testing.T) {
	uuid_ := uuid.New()
	time_ := time.Now()
	tests := []struct {
		name    string
		id      uuid.UUID
		mock    func(mock sqlmock.Sqlmock)
		want    models.Community
		wantErr bool
	}{
		{
			name: "Successfully get community by id",
			id:   uuid_,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`select id, owner_id, name, description`).WithArgs(uuid_).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "owner_id", "name", "description", "created_at", "avatar_url", "cover_url", "contact_info", "nickname",
					}).
						AddRow(uuid_, uuid_, "Test Community", "Description of community", time_, "http://avatar.url", "http://cover.url", nil, "TestCommunity"))
			},
			want: models.Community{
				ID:        uuid_,
				OwnerID:   uuid_,
				NickName:  "TestCommunity",
				CreatedAt: time_,
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Test Community",
					Description: "Description of community",
					AvatarUrl:   "http://avatar.url",
					CoverUrl:    "http://cover.url",
				},
			},
			wantErr: false,
		},
		{
			name: "Failed to get community by id",
			id:   uuid_,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`select id, owner_id, name, description`).WithArgs(uuid_).
					WillReturnError(fmt.Errorf("select failed"))
			},
			want:    models.Community{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			repo := &SqlCommunityRepository{connPool: mockDB}
			tt.mock(mock)

			got, err := repo.GetCommunityById(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommunityById() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Compare times using `Equal` to avoid failure due to small differences
			if !got.CreatedAt.Equal(tt.want.CreatedAt) {
				t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, tt.want.CreatedAt)
			}

			// Compare all fields except time
			tt.want.CreatedAt = got.CreatedAt
			assert.Equal(t, tt.want, got)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdateCommunityTextInfo(t *testing.T) {
	tests := []struct {
		name      string
		community models.Community
		mock      func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "Successfully update community",
			community: models.Community{
				ID:       uuid.New(),
				NickName: "UpdatedCommunity",
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Updated Name",
					Description: "Updated Description",
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`update community set nickname`).WithArgs(
					"UpdatedCommunity", "Updated Name", "Updated Description", sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Failed to update community",
			community: models.Community{
				ID:       uuid.New(),
				NickName: "UpdatedCommunity",
				BasicInfo: &models.BasicCommunityInfo{
					Name:        "Updated Name",
					Description: "Updated Description",
				},
			},
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`update community set nickname`).WithArgs(
					"UpdatedCommunity", "Updated Name", "Updated Description", sqlmock.AnyArg()).
					WillReturnError(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			repo := &SqlCommunityRepository{connPool: mockDB}
			tt.mock(mock)

			err = repo.UpdateCommunityTextInfo(context.Background(), tt.community)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCommunityTextInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
