package get_env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServiceAddr(t *testing.T) {
	tests := []struct {
		name               string
		runningInContainer string
		serviceVar         string
		serviceEnvValue    string
		defaultPort        int
		expectedAddr       string
	}{
		{
			name:               "Test container environment with service address",
			runningInContainer: "true",
			serviceVar:         "SERVICE_ADDR",
			serviceEnvValue:    "192.168.1.100:8080",
			defaultPort:        8080,
			expectedAddr:       "192.168.1.100:8080",
		},
		{
			name:               "Test container environment without service address",
			runningInContainer: "true",
			serviceVar:         "SERVICE_ADDR",
			serviceEnvValue:    "",
			defaultPort:        8080,
			expectedAddr:       "127.0.0.1:8080",
		},
		{
			name:               "Test non-container environment",
			runningInContainer: "false",
			serviceVar:         "SERVICE_ADDR",
			serviceEnvValue:    "",
			defaultPort:        8080,
			expectedAddr:       "127.0.0.1:8080",
		},
		{
			name:               "Test non-container environment with fallback service address",
			runningInContainer: "false",
			serviceVar:         "SERVICE_ADDR",
			serviceEnvValue:    "192.168.1.100:9090",
			defaultPort:        8080,
			expectedAddr:       "127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("RUNNING_IN_CONTAINER", tt.runningInContainer)
			os.Setenv(tt.serviceVar, tt.serviceEnvValue)

			actualAddr := GetServiceAddr(tt.serviceVar, tt.defaultPort)
			assert.Equal(t, tt.expectedAddr, actualAddr)
		})
	}
}
