package models

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateSession(t *testing.T) {

	tests := []struct {
		name       string
		ExpireDate time.Time
	}{
		{
			name:       "Test CreateSession",
			ExpireDate: time.Now().Add(10 * 24 * time.Hour),
		},
	}

	for _, test := range tests {
		session := CreateSession()
		if session.ExpireDate.Date() != test.ExpireDate.Date() || session.SessionId == uuid.Nil {
			assert.Fail(t, "Test failed: %s", test.name)
		}
	}
}
