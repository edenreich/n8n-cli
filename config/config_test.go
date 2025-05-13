package config_test

import (
	"testing"

	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/config/configfakes"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestLoadConfig is a unit test that verifies LoadConfig behavior
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		instanceURL   string
		expectedError bool
	}{
		{
			name:          "Valid configuration",
			apiKey:        "test-api-key",
			instanceURL:   "http://test-url:5678",
			expectedError: false,
		},
		{
			name:          "Missing API key",
			apiKey:        "",
			instanceURL:   "http://test-url:5678",
			expectedError: true,
		},
		{
			name:          "Missing instance URL",
			apiKey:        "test-api-key",
			instanceURL:   "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new viper instance for isolation
			v := viper.New()
			v.Set("api_key", tt.apiKey)
			v.Set("instance_url", tt.instanceURL)

			cfg, err := config.LoadConfig(v)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, tt.apiKey, cfg.APIToken)
				assert.Equal(t, tt.instanceURL+"/api/v1", cfg.APIBaseURL)
			}
		})
	}
}

// TestConfigFakes demonstrates how to use generated fake config interface
func TestConfigFakes(t *testing.T) {
	// Create a fake config instance
	fakeConfig := &configfakes.FakeConfigInterface{}

	// Configure the fake to return specific values
	fakeConfig.GetAPITokenReturns("fake-api-token")
	fakeConfig.GetAPIBaseURLReturns("http://fake-url:5678/api/v1")

	// Use the fake config in your test
	assert.Equal(t, "fake-api-token", fakeConfig.GetAPIToken())
	assert.Equal(t, "http://fake-url:5678/api/v1", fakeConfig.GetAPIBaseURL())

	// Verify that the methods were called
	assert.Equal(t, 1, fakeConfig.GetAPITokenCallCount())
	assert.Equal(t, 1, fakeConfig.GetAPIBaseURLCallCount())
}

// TestConfigWithMockInUse demonstrates how to use the fake config in a function
// that expects a ConfigInterface
func TestConfigWithMockInUse(t *testing.T) {
	// Create a fake config
	fakeConfig := &configfakes.FakeConfigInterface{}
	fakeConfig.GetAPITokenReturns("mock-token")
	fakeConfig.GetAPIBaseURLReturns("http://mock-api:5678")

	// This function simulates a function that would use the config interface
	getFullAPIURL := func(cfg config.ConfigInterface, endpoint string) string {
		return cfg.GetAPIBaseURL() + endpoint
	}

	// Test the function with our fake config
	endpoint := "/workflows"
	result := getFullAPIURL(fakeConfig, endpoint)

	// Verify expectations
	assert.Equal(t, "http://mock-api:5678/workflows", result)
	assert.Equal(t, 1, fakeConfig.GetAPIBaseURLCallCount())
}
