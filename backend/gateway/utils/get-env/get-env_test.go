package get_env_test

import (
	"os"
	get_env "quickflow/utils/get-env"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	const key = "TEST_GET_ENV"

	tests := []struct {
		name         string
		setValue     *string
		defaultValue string
		expected     string
	}{
		{
			name:         "env variable is set",
			setValue:     strPtr("hello"),
			defaultValue: "default",
			expected:     "hello",
		},
		{
			name:         "env variable is not set",
			setValue:     nil,
			defaultValue: "fallback",
			expected:     "fallback",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setValue != nil {
				os.Setenv(key, *test.setValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}

			actual := get_env.GetEnv(key, test.defaultValue)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	const key = "TEST_GET_ENV_INT"

	tests := []struct {
		name         string
		setValue     *string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid int value",
			setValue:     strPtr("42"),
			defaultValue: 0,
			expected:     42,
		},
		{
			name:         "invalid int value",
			setValue:     strPtr("notanint"),
			defaultValue: 100,
			expected:     100,
		},
		{
			name:         "unset variable",
			setValue:     nil,
			defaultValue: 7,
			expected:     7,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setValue != nil {
				os.Setenv(key, *test.setValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}

			actual := get_env.GetEnvAsInt(key, test.defaultValue)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	const key = "TEST_GET_ENV_BOOL"

	tests := []struct {
		name         string
		setValue     *string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "true value",
			setValue:     strPtr("true"),
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "false value",
			setValue:     strPtr("false"),
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "invalid bool",
			setValue:     strPtr("yes"),
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "unset value",
			setValue:     nil,
			defaultValue: false,
			expected:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setValue != nil {
				os.Setenv(key, *test.setValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}

			actual := get_env.GetEnvAsBool(key, test.defaultValue)
			require.Equal(t, test.expected, actual)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
