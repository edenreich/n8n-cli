package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	originalAPIKey := viper.GetString("api_key")
	originalInstanceURL := viper.GetString("instance_url")
	defer func() {
		viper.Set("api_key", originalAPIKey)
		viper.Set("instance_url", originalInstanceURL)
	}()

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
			viper.Set("api_key", tt.apiKey)
			viper.Set("instance_url", tt.instanceURL)

			cfg, err := GetConfig()

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

func TestConfigFromEnvVars(t *testing.T) {
	origAPIKey := os.Getenv("N8N_API_KEY")
	origInstanceURL := os.Getenv("N8N_INSTANCE_URL")

	defer func() {
		err := os.Setenv("N8N_API_KEY", origAPIKey)
		assert.NoError(t, err)
		err = os.Setenv("N8N_INSTANCE_URL", origInstanceURL)
		assert.NoError(t, err)

		viper.Reset()
	}()

	viper.Reset()

	err := os.Setenv("N8N_API_KEY", "env-test-api-key")
	assert.NoError(t, err)
	err = os.Setenv("N8N_INSTANCE_URL", "http://env-test-url:5678")
	assert.NoError(t, err)

	viper.SetEnvPrefix("N8N")
	viper.AutomaticEnv()
	err = viper.BindEnv("api_key", "N8N_API_KEY")
	assert.NoError(t, err)
	err = viper.BindEnv("instance_url", "N8N_INSTANCE_URL")
	assert.NoError(t, err)

	viper.SetDefault("instance_url", "http://localhost:5678")
	viper.SetDefault("api_key", "")

	cfg, err := GetConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "env-test-api-key", cfg.APIToken)
	assert.Equal(t, "http://env-test-url:5678/api/v1", cfg.APIBaseURL)
}

func TestConfigPrecedence(t *testing.T) {
	origAPIKey := os.Getenv("N8N_API_KEY")
	origInstanceURL := os.Getenv("N8N_INSTANCE_URL")
	origViperAPIKey := viper.GetString("api_key")
	origViperURL := viper.GetString("instance_url")
	defer func() {
		err := os.Setenv("N8N_API_KEY", origAPIKey)
		assert.NoError(t, err)
		err = os.Setenv("N8N_INSTANCE_URL", origInstanceURL)
		assert.NoError(t, err)
		viper.Set("api_key", origViperAPIKey)
		viper.Set("instance_url", origViperURL)
	}()

	err := os.Setenv("N8N_API_KEY", "env-api-key")
	assert.NoError(t, err)
	err = os.Setenv("N8N_INSTANCE_URL", "http://env-url:5678")
	assert.NoError(t, err)

	viper.Set("api_key", "flag-api-key")
	viper.Set("instance_url", "http://flag-url:5678")

	viper.SetEnvPrefix("N8N")
	viper.AutomaticEnv()
	err = viper.BindEnv("api_key", "N8N_API_KEY")
	assert.NoError(t, err)
	err = viper.BindEnv("instance_url", "N8N_INSTANCE_URL")
	assert.NoError(t, err)

	cfg, err := GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, "flag-api-key", cfg.APIToken)
	assert.Equal(t, "http://flag-url:5678/api/v1", cfg.APIBaseURL)
}
