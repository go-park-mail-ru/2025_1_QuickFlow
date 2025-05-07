package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"quickflow/file_service/config/validation"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		name      string
		sizeStr   string
		expected  int64
		expectErr bool
	}{
		{
			name:     "Valid KB size",
			sizeStr:  "10KB",
			expected: 10240, // 10 * 1024
		},
		{
			name:     "Valid MB size",
			sizeStr:  "5MB",
			expected: 5242880, // 5 * 1024 * 1024
		},
		{
			name:     "Valid GB size",
			sizeStr:  "2GB",
			expected: 2147483648, // 2 * 1024 * 1024 * 1024
		},
		{
			name:     "Valid B size",
			sizeStr:  "500B",
			expected: 500,
		},
		{
			name:      "Invalid size format",
			sizeStr:   "100XYZ",
			expected:  0,
			expectErr: true,
		},
		{
			name:     "Valid size without unit",
			sizeStr:  "1000",
			expected: 1000, // Assuming it's in bytes
		},
		{
			name:     "Valid mixed case size",
			sizeStr:  "1gB",
			expected: 1073741824, // 1 * 1024 * 1024 * 1024
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validation.ParseSize(tt.sizeStr)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
