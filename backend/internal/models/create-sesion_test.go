package models

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateSession(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			name: "Test CreateSession",
		},
	}

	for _, test := range tests {
		session := CreateSession()
		if session.SessionId == uuid.Nil && !session.ExpireDate.IsZero() {
			assert.Fail(t, "Test failed: %s", test.name)
		}
	}
}
