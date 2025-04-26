package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateFeedParams(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		numPosts  int
		timestamp time.Time
		expected  error
	}{
		{
			name:      "valid parameters",
			numPosts:  10,
			timestamp: now.Add(-time.Hour), // в прошлом
			expected:  nil,
		},
		{
			name:      "zero posts",
			numPosts:  0,
			timestamp: now.Add(-time.Hour),
			expected:  ErrInvalidNumPosts,
		},
		{
			name:      "negative posts",
			numPosts:  -5,
			timestamp: now.Add(-time.Hour),
			expected:  ErrInvalidNumPosts,
		},
		{
			name:      "timestamp is zero",
			numPosts:  5,
			timestamp: time.Time{},
			expected:  ErrInvalidTimestamp,
		},
		{
			name:      "timestamp is in the future",
			numPosts:  5,
			timestamp: now.Add(time.Hour),
			expected:  ErrInvalidTimestamp,
		},
	}

	for _, tt := range tests {
		err := ValidateFeedParams(tt.numPosts, tt.timestamp)
		if tt.expected != nil {
			require.ErrorIs(t, err, tt.expected, tt.name)
		} else {
			require.NoError(t, err, tt.name)
		}
	}
}
