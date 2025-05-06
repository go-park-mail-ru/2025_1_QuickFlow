package forms_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"quickflow/gateway/internal/delivery/forms"
)

func TestSearchForm_Unpack(t *testing.T) {
	tests := []struct {
		name        string
		values      url.Values
		expected    forms.SearchForm
		expectedErr error
	}{
		{
			name: "Success - all fields present",
			values: url.Values{
				"string": []string{"testuser"},
				"count":  []string{"10"},
			},
			expected: forms.SearchForm{
				ToSearch: "testuser",
				Count:    10,
			},
			expectedErr: nil,
		},
		{
			name: "Missing string parameter",
			values: url.Values{
				"count": []string{"10"},
			},
			expected:    forms.SearchForm{},
			expectedErr: errors.New("username parameter missing"),
		},
		{
			name: "Missing count parameter",
			values: url.Values{
				"string": []string{"testuser"},
			},
			expected:    forms.SearchForm{},
			expectedErr: errors.New("count parameter missing"),
		},
		{
			name: "Invalid count format",
			values: url.Values{
				"string": []string{"testuser"},
				"count":  []string{"notanumber"},
			},
			expected:    forms.SearchForm{},
			expectedErr: errors.New("failed to parse count"),
		},
		{
			name: "Empty string parameter",
			values: url.Values{
				"string": []string{""},
				"count":  []string{"10"},
			},
			expected: forms.SearchForm{
				ToSearch: "",
				Count:    10,
			},
			expectedErr: nil,
		},
		{
			name: "Count zero value",
			values: url.Values{
				"string": []string{"testuser"},
				"count":  []string{"0"},
			},
			expected: forms.SearchForm{
				ToSearch: "testuser",
				Count:    0,
			},
			expectedErr: nil,
		},
		{
			name: "Negative count value",
			values: url.Values{
				"string": []string{"testuser"},
				"count":  []string{"-5"},
			},
			expected:    forms.SearchForm{},
			expectedErr: errors.New("count must be positive"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var form forms.SearchForm
			err := form.Unpack(tt.values)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, form)
			}
		})
	}
}
