package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"quickflow/shared/models"
)

func TestSaveSession(t *testing.T) {
	tests := []struct {
		name    string
		userId  uuid.UUID
		session models.Session
		mock    func(mock redismock.ClientMock)
		wantErr bool
	}{
		{
			name:   "Successfully save session",
			userId: uuid.MustParse("9e49c172-8626-4c60-8240-6b8e774e0a4a"),
			session: models.Session{
				SessionId:  uuid.MustParse("22896b51-8736-42dc-bf6f-b438c1ad3aa5"),
				ExpireDate: time.Now().Add(24 * time.Hour),
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectSet("22896b51-8736-42dc-bf6f-b438c1ad3aa5", "9e49c172-8626-4c60-8240-6b8e774e0a4a", time.Until(time.Now().Add(24*time.Hour))).
					SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:   "Failed to save session",
			userId: uuid.MustParse("331b7880-4f48-4925-a312-67a848e631f2"),
			session: models.Session{
				SessionId:  uuid.MustParse("5998eccb-91e1-40a5-8b02-9883cb0ac95d"),
				ExpireDate: time.Now().Add(24 * time.Hour),
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectSet("5998eccb-91e1-40a5-8b02-9883cb0ac95d", "331b7880-4f48-4925-a312-67a848e631f2", time.Until(time.Now().Add(24*time.Hour))).
					SetErr(fmt.Errorf("failed to save"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock Redis client
			mockDB, mock := redismock.NewClientMock()
			tt.mock(mock)

			repo := &RedisSessionRepository{rdb: mockDB}

			err := repo.SaveSession(context.Background(), tt.userId, tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLookupUserSession(t *testing.T) {
	tests := []struct {
		name    string
		session models.Session
		mock    func(mock redismock.ClientMock)
		want    uuid.UUID
		wantErr bool
	}{
		{
			name: "Successfully lookup user session",
			session: models.Session{
				SessionId: uuid.MustParse("22896b51-8736-42dc-bf6f-b438c1ad3aa5"),
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("22896b51-8736-42dc-bf6f-b438c1ad3aa5").
					SetVal("9e49c172-8626-4c60-8240-6b8e774e0a4a")
			},
			want:    uuid.MustParse("9e49c172-8626-4c60-8240-6b8e774e0a4a"),
			wantErr: false,
		},
		{
			name: "Failed to lookup user session",
			session: models.Session{
				SessionId: uuid.MustParse("22896b51-8736-42dc-bf6f-b438c1ad3aa5"),
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("22896b51-8736-42dc-bf6f-b438c1ad3aa5").
					SetErr(fmt.Errorf("failed to find"))
			},
			want:    uuid.Nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock Redis client
			mockDB, mock := redismock.NewClientMock()
			tt.mock(mock)

			repo := &RedisSessionRepository{rdb: mockDB}

			got, err := repo.LookupUserSession(context.Background(), tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookupUserSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsExists(t *testing.T) {
	tests := []struct {
		name    string
		session uuid.UUID
		mock    func(mock redismock.ClientMock)
		want    bool
		wantErr bool
	}{
		{
			name:    "Session exists",
			session: uuid.MustParse("22896b51-8736-42dc-bf6f-b438c1ad3aa5"),
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("22896b51-8736-42dc-bf6f-b438c1ad3aa5").
					SetVal("9e49c172-8626-4c60-8240-6b8e774e0a4a")
			},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Session does not exist",
			session: uuid.MustParse("22896b51-8736-42dc-bf6f-b438c1ad3aa5"),
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("22896b51-8736-42dc-bf6f-b438c1ad3aa5").
					SetErr(redis.Nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:    "Failed to check session",
			session: uuid.MustParse("22896b51-8736-42dc-bf6f-b438c1ad3aa5"),
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("22896b51-8736-42dc-bf6f-b438c1ad3aa5").
					SetErr(fmt.Errorf("failed to check"))
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock Redis client
			mockDB, mock := redismock.NewClientMock()
			tt.mock(mock)

			repo := &RedisSessionRepository{rdb: mockDB}

			got, err := repo.IsExists(context.Background(), tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		name    string
		session string
		mock    func(mock redismock.ClientMock)
		wantErr bool
	}{
		{
			name:    "Successfully delete session",
			session: "sessionId",
			mock: func(mock redismock.ClientMock) {
				mock.ExpectDel("sessionId").
					SetVal(1)
			},
			wantErr: false,
		},
		{
			name:    "Failed to delete session",
			session: "sessionId",
			mock: func(mock redismock.ClientMock) {
				mock.ExpectDel("sessionId").
					SetErr(fmt.Errorf("failed to delete"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock Redis client
			mockDB, mock := redismock.NewClientMock()
			tt.mock(mock)

			repo := &RedisSessionRepository{rdb: mockDB}

			err := repo.DeleteSession(context.Background(), tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
