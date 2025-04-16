package postgres

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"quickflow/internal/models"
	"testing"
)

func TestGetFriendsPublicInfo(t *testing.T) {
	uuid_ := uuid.New()
	tests := []struct {
		name        string
		userID      string
		limit       int
		offset      int
		mock        func(mock sqlmock.Sqlmock)
		want        []models.FriendInfo
		wantHasMore bool
		wantCount   int
		wantErr     bool
	}{
		{
			name:   "Successfully get friends info",
			userID: "user1",
			limit:  5,
			offset: 0,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`with friends as \(.*\)`).
					WithArgs("user1", 6, 0, models.RelationFriend).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "firstname", "lastname", "profile_avatar", "name"}).
						AddRow(uuid_, "johndoe", "John", "Doe", "http://avatar.url", "Some University"))

				mock.ExpectQuery(`select count\(\*\)`).
					WithArgs("user1", models.RelationFriend).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

			},
			want: []models.FriendInfo{
				{
					Id:         uuid_,
					Username:   "johndoe",
					Firstname:  "John",
					Lastname:   "Doe",
					AvatarURL:  "http://avatar.url",
					University: "Some University",
				},
			},
			wantHasMore: false,
			wantCount:   1,
			wantErr:     false,
		},
		{
			name:   "Failed to get friends info",
			userID: "user1",
			limit:  5,
			offset: 0,
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`with friends as \(.*\)`).
					WithArgs("user1", 6, 0, models.RelationFriend).
					WillReturnError(fmt.Errorf("query failed"))
			},
			want:        []models.FriendInfo{},
			wantHasMore: false,
			wantCount:   0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to open mock DB: %v", err)
			}
			defer mockDB.Close()

			repo := &PostgresFriendsRepository{connPool: mockDB}
			tt.mock(mock)

			got, hasMore, count, err := repo.GetFriendsPublicInfo(context.Background(), tt.userID, tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFriendsPublicInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantHasMore, hasMore)
			assert.Equal(t, tt.wantCount, count)

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestSendFriendRequest(t *testing.T) {
	tests := []struct {
		name       string
		senderID   string
		receiverID string
		mock       func(mock sqlmock.Sqlmock)
		wantErr    bool
	}{
		{
			name:       "Successfully send friend request",
			senderID:   "user1",
			receiverID: "user2",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`insert into friendship`).
					WithArgs("user1", "user2", models.RelationFollowing).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:       "Failed to send friend request",
			senderID:   "user1",
			receiverID: "user2",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`insert into friendship`).
					WithArgs("user1", "user2", models.RelationFollowing).
					WillReturnError(fmt.Errorf("query failed"))
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

			repo := &PostgresFriendsRepository{connPool: mockDB}
			tt.mock(mock)

			err = repo.SendFriendRequest(context.Background(), tt.senderID, tt.receiverID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendFriendRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestDeleteFriend(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		friendID string
		mock     func(mock sqlmock.Sqlmock)
		wantErr  bool
	}{
		{
			name:     "Successfully delete friend",
			userID:   "user1",
			friendID: "user2",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`update friendship`).
					WithArgs("user1", "user2", models.RelationFollowedBy, models.RelationFriend).
					WillReturnResult(sqlmock.NewResult(1, 1))

			},
			wantErr: false,
		},
		{
			name:     "Failed to delete friend",
			userID:   "user1",
			friendID: "user2",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`update friendship`).
					WithArgs("user1", "user2", models.RelationFollowedBy, models.RelationFriend).
					WillReturnError(fmt.Errorf("query failed"))

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

			repo := &PostgresFriendsRepository{connPool: mockDB}
			tt.mock(mock)

			err = repo.DeleteFriend(context.Background(), tt.userID, tt.friendID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteFriend() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUnfollow(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		friendID string
		mock     func(mock sqlmock.Sqlmock)
		wantErr  bool
	}{
		{
			name:     "Successfully unfollow",
			userID:   "user1",
			friendID: "user2",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`delete from friendship`).
					WithArgs("user1", "user2", models.RelationFollowedBy, models.RelationFollowing).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:     "Failed to unfollow",
			userID:   "user1",
			friendID: "user2",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`delete from friendship`).
					WithArgs("user1", "user2", models.RelationFollowedBy, models.RelationFollowing).
					WillReturnError(fmt.Errorf("query failed"))
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

			repo := &PostgresFriendsRepository{connPool: mockDB}
			tt.mock(mock)

			err = repo.Unfollow(context.Background(), tt.userID, tt.friendID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unfollow() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Ensure that all expected queries were executed
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
