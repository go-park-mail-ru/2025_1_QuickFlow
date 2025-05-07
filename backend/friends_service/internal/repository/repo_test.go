package postgres_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"quickflow/friends_service/internal/repository"
	"quickflow/shared/models"
)

func TestGetFriendsPublicInfo(t *testing.T) {
	// Initialize mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("could not open db mock: %v", err)
	}
	defer db.Close()

	repo := postgres.NewPostgresFriendsRepository(db)
	userID := "user-123"
	limit := 10
	offset := 0
	reqType := "all"

	// Case 1: Valid query with friends
	rows := sqlmock.NewRows([]string{"id", "username", "firstname", "lastname", "profile_avatar", "university"}).
		AddRow(uuid.New(), "user1", "John", "Doe", "avatar1.jpg", "University A").
		AddRow(uuid.New(), "user2", "Jane", "Doe", "avatar2.jpg", "University B")

	mock.ExpectQuery(`with related_users as \( select case when user1_id = \$1 then user2_id else user1_id end as related_id from friendship where\(\(user1_id = \$1 AND status = \$2\) OR \(user2_id = \$1 AND status = \$3\)\) \) select u.id, u.username, p.firstname, p.lastname, p.profile_avatar, univ.name from "user" u join profile p on u.id = p.id left join education e on e.profile_id = p.id left join faculty f on f.id = e.faculty_id left join university univ on f.university_id = univ.id where u.id in \(select related_id from related_users\) order by p.lastname, p.firstname limit \$4 offset \$5`).
		WithArgs(userID, models.RelationFriend, models.RelationFriend, limit, offset).
		WillReturnRows(rows)

	// Case 2: Valid friends count query
	mock.ExpectQuery(`select count\( case when user1_id = \$1 then user2_id else user1_id end \) from friendship where \(\(user1_id = \$1 and status = \$2\) or \(user2_id = \$1 and status = \$3\)\)`).
		WithArgs(userID, models.RelationFriend, models.RelationFriend).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Call the method under test
	friendsInfo, friendsCount, err := repo.GetFriendsPublicInfo(context.Background(), userID, limit, offset, reqType)

	// Assert the result
	require.NoError(t, err)
	assert.Equal(t, 2, friendsCount)
	assert.Len(t, friendsInfo, 2)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestSendFriendRequest(t *testing.T) {
	// Initialize mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("could not open db mock: %v", err)
	}
	defer db.Close()

	repo := postgres.NewPostgresFriendsRepository(db)
	senderID := "user-123"
	receiverID := "user-456"

	// Case 1: Valid request insertion
	mock.ExpectExec(`insert into friendship \(user1_id, user2_id, status\) values \(\$1, \$2, \$3\)`).
		WithArgs(senderID, receiverID, models.RelationFollowing).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method under test
	err = repo.SendFriendRequest(context.Background(), senderID, receiverID)

	// Assert the result
	assert.NoError(t, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestIsExistsFriendRequest(t *testing.T) {
	// Initialize mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("could not open db mock: %v", err)
	}
	defer db.Close()

	repo := postgres.NewPostgresFriendsRepository(db)
	senderID := "user-123"
	receiverID := "user-456"

	// Case 1: Friend request exists
	mock.ExpectQuery(`select status from friendship where \(user1_id = \$1 and user2_id = \$2\) or \(user1_id = \$2 and user2_id = \$1\)`).
		WithArgs(senderID, receiverID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.RelationFollowing))

	// Call the method under test
	exists, err := repo.IsExistsFriendRequest(context.Background(), senderID, receiverID)

	// Assert the result
	require.NoError(t, err)
	assert.True(t, exists)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAcceptFriendRequest(t *testing.T) {
	// Initialize mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("could not open db mock: %v", err)
	}
	defer db.Close()

	repo := postgres.NewPostgresFriendsRepository(db)
	senderID := "user-123"
	receiverID := "user-456"

	// Case 1: Valid accept request
	mock.ExpectExec(`update friendship set status = \$3 where user1_id = \$1 and user2_id = \$2 and status != \$3`).
		WithArgs(senderID, receiverID, models.RelationFriend).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method under test
	err = repo.AcceptFriendRequest(context.Background(), senderID, receiverID)

	// Assert the result
	assert.NoError(t, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestUnfollow(t *testing.T) {
	// Initialize mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("could not open db mock: %v", err)
	}
	defer db.Close()

	repo := postgres.NewPostgresFriendsRepository(db)
	userID := "user-123"
	friendID := "user-456"

	// Case 1: Valid unfollow
	mock.ExpectExec(`delete from friendship where \(\(user1_id = \$1 and user2_id = \$2\) or \(user1_id = \$2 and user2_id = \$1\)\) and status in \(\$3, \$4\)`).
		WithArgs(userID, friendID, models.RelationFollowedBy, models.RelationFollowing).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method under test
	err = repo.Unfollow(context.Background(), userID, friendID)

	// Assert the result
	assert.NoError(t, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
