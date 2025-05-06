package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateReqType(t *testing.T) {
	// Table-driven test cases
	tests := []struct {
		name     string
		reqType  string
		expected bool
	}{
		{
			name:     "Valid 'outcoming' type",
			reqType:  TypeOutComing,
			expected: true,
		},
		{
			name:     "Valid 'incoming' type",
			reqType:  TypeInComing,
			expected: true,
		},
		{
			name:     "Valid 'all' type",
			reqType:  TypeAll,
			expected: true,
		},
		{
			name:     "Invalid 'invalid' type",
			reqType:  "invalid",
			expected: false,
		},
		{
			name:     "Empty type",
			reqType:  "",
			expected: false,
		},
	}

	// Iterate over the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function with the test case input
			actual := validateReqType(tt.reqType)

			// Assert that the result matches the expected value
			assert.Equal(t, tt.expected, actual)
		})
	}
}
