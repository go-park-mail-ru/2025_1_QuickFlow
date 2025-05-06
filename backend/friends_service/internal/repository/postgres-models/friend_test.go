package postgres_models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestConvertToFriendInfo(t *testing.T) {
	// Prepare the FriendInfoPostgres struct
	id := uuid.New()
	friendInfoPostgres := &FriendInfoPostgres{
		Id:         pgtype.UUID{Bytes: [16]byte(id[:])}, // Convert UUID to [16]byte
		Username:   pgtype.Text{String: "user1", Valid: true},
		Firstname:  pgtype.Text{String: "John", Valid: true},
		Lastname:   pgtype.Text{String: "Doe", Valid: true},
		AvatarURL:  pgtype.Text{String: "avatar1.jpg", Valid: true},
		University: pgtype.Text{String: "University A", Valid: true},
	}

	// Convert to FriendInfo
	convertedFriendInfo := friendInfoPostgres.ConvertToFriendInfo()

	// Convert [16]byte to uuid.UUID using the uuid.UUID constructor
	convertedID := uuid.UUID(friendInfoPostgres.Id.Bytes)

	// Assert the conversion is correct
	assert.Equal(t, id, convertedID) // Compare UUIDs
	assert.Equal(t, "user1", convertedFriendInfo.Username)
	assert.Equal(t, "John", convertedFriendInfo.Firstname)
	assert.Equal(t, "Doe", convertedFriendInfo.Lastname)
	assert.Equal(t, "avatar1.jpg", convertedFriendInfo.AvatarURL)
	assert.Equal(t, "University A", convertedFriendInfo.University)
}
