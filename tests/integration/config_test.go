// Package integration contains integration tests for the n8n-cli
package integration

import (
	"os"
	"testing"

	"github.com/edenreich/n8n-cli/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestConfigFromEnvVars(t *testing.T) {
	origAPIKey := os.Getenv("N8N_API_KEY")
	origInstanceURL := os.Getenv("N8N_INSTANCE_URL")

	defer func() {
		err := os.Setenv("N8N_API_KEY", origAPIKey)
		assert.NoError(t, err)
		err = os.Setenv("N8N_INSTANCE_URL", origInstanceURL)
		assert.NoError(t, err)
	}()

	err := os.Setenv("N8N_API_KEY", "env-test-api-key")
	assert.NoError(t, err)
	err = os.Setenv("N8N_INSTANCE_URL", "http://env-test-url:5678")
	assert.NoError(t, err)

	v := viper.New()
	v.SetEnvPrefix("N8N")
	v.AutomaticEnv()
	err = v.BindEnv("api_key", "N8N_API_KEY")
	assert.NoError(t, err)
	err = v.BindEnv("instance_url", "N8N_INSTANCE_URL")
	assert.NoError(t, err)

	v.SetDefault("instance_url", "http://localhost:5678")
	v.SetDefault("api_key", "")

	cfg, err := config.LoadConfig(v)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "env-test-api-key", cfg.APIToken)
	assert.Equal(t, "http://env-test-url:5678/api/v1", cfg.APIBaseURL)
}

func TestConfigPrecedence(t *testing.T) {
	origAPIKey := os.Getenv("N8N_API_KEY")
	origInstanceURL := os.Getenv("N8N_INSTANCE_URL")

	defer func() {
		err := os.Setenv("N8N_API_KEY", origAPIKey)
		assert.NoError(t, err)
		err = os.Setenv("N8N_INSTANCE_URL", origInstanceURL)
		assert.NoError(t, err)
	}()

	err := os.Setenv("N8N_API_KEY", "env-api-key")
	assert.NoError(t, err)
	err = os.Setenv("N8N_INSTANCE_URL", "http://env-url:5678")
	assert.NoError(t, err)

	v := viper.New()
	v.Set("api_key", "flag-api-key")
	v.Set("instance_url", "http://flag-url:5678")

	v.SetEnvPrefix("N8N")
	v.AutomaticEnv()
	err = v.BindEnv("api_key", "N8N_API_KEY")
	assert.NoError(t, err)
	err = v.BindEnv("instance_url", "N8N_INSTANCE_URL")
	assert.NoError(t, err)

	cfg, err := config.LoadConfig(v)
	assert.NoError(t, err)
	assert.Equal(t, "flag-api-key", cfg.APIToken)
	assert.Equal(t, "http://flag-url:5678/api/v1", cfg.APIBaseURL)
}
