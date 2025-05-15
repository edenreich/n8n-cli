package unit

import (
	"os"
	"testing"

	"github.com/edenreich/n8n-cli/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestLoadConfig is a unit test that verifies LoadConfig behavior
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		instanceURL string
		wantKey     string
		wantURL     string
	}{
		{
			name:        "Valid configuration",
			apiKey:      "test-api-key",
			instanceURL: "http://test-url:5678",
			wantKey:     "test-api-key",
			wantURL:     "http://test-url:5678",
		},
		{
			name:        "Missing API key",
			apiKey:      "",
			instanceURL: "http://test-url:5678",
			wantKey:     "",
			wantURL:     "http://test-url:5678",
		},
		{
			name:        "Missing instance URL",
			apiKey:      "test-api-key",
			instanceURL: "",
			wantKey:     "test-api-key",
			wantURL:     "http://localhost:5678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Unsetenv("N8N_API_KEY")
			assert.NoError(t, err)
			err = os.Unsetenv("N8N_INSTANCE_URL")
			assert.NoError(t, err)
			viper.Reset()

			if tt.apiKey != "" {
				err = os.Setenv("N8N_API_KEY", tt.apiKey)
				assert.NoError(t, err)
			}
			if tt.instanceURL != "" {
				err = os.Setenv("N8N_INSTANCE_URL", tt.instanceURL)
				assert.NoError(t, err)
			}

			config.Initialize()

			assert.Equal(t, tt.wantKey, viper.GetString("api_key"))
			assert.Equal(t, tt.wantURL, viper.GetString("instance_url"))
		})
	}
}

// TestBindEnvSafely tests that environment variables are properly bound
func TestBindEnvSafely(t *testing.T) {
	err := os.Setenv("N8N_TEST_VAR", "test-value")
	assert.NoError(t, err)
	defer func() {
		err := os.Unsetenv("N8N_TEST_VAR")
		assert.NoError(t, err)
	}()

	v := viper.New()
	config.BindEnvSafely(v, "test_var", "N8N_TEST_VAR")

	assert.Equal(t, "test-value", v.GetString("test_var"))
}

// TestInitializeWithConfigFile tests loading from a config file
func TestInitializeWithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
api_key: config-file-key
instance_url: http://config-file-url:5678
`
	err := os.WriteFile(tmpDir+"/config.yaml", []byte(configContent), 0644)
	assert.NoError(t, err)

	err = os.Unsetenv("N8N_API_KEY")
	assert.NoError(t, err)
	err = os.Unsetenv("N8N_INSTANCE_URL")
	assert.NoError(t, err)
	viper.Reset()

	viper.AddConfigPath(tmpDir)

	config.Initialize()

	assert.Equal(t, "config-file-key", viper.GetString("api_key"))
	assert.Equal(t, "http://config-file-url:5678", viper.GetString("instance_url"))
}

// TestEnvOverridesConfigFile verifies that environment variables take precedence over config file values
func TestEnvOverridesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
api_key: config-file-key
instance_url: http://config-file-url:5678
`
	err := os.WriteFile(tmpDir+"/config.yaml", []byte(configContent), 0644)
	assert.NoError(t, err)

	viper.Reset()

	err = os.Setenv("N8N_API_KEY", "env-api-key")
	assert.NoError(t, err)
	err = os.Setenv("N8N_INSTANCE_URL", "http://env-url:5678")
	assert.NoError(t, err)
	defer func() {
		err := os.Unsetenv("N8N_API_KEY")
		assert.NoError(t, err)
		err = os.Unsetenv("N8N_INSTANCE_URL")
		assert.NoError(t, err)
	}()

	viper.AddConfigPath(tmpDir)

	config.Initialize()

	assert.Equal(t, "env-api-key", viper.GetString("api_key"))
	assert.Equal(t, "http://env-url:5678", viper.GetString("instance_url"))
}

// TestDefaultValues verifies that default values are set correctly when no config is provided
func TestDefaultValues(t *testing.T) {
	err := os.Unsetenv("N8N_API_KEY")
	assert.NoError(t, err)
	err = os.Unsetenv("N8N_INSTANCE_URL")
	assert.NoError(t, err)
	viper.Reset()

	config.Initialize()

	assert.Equal(t, "", viper.GetString("api_key"))
	assert.Equal(t, "http://localhost:5678", viper.GetString("instance_url"))
}

// TestBindEnvSafelyErrorHandling verifies that the function doesn't crash on invalid inputs
func TestBindEnvSafelyErrorHandling(t *testing.T) {
	v := viper.New()

	config.BindEnvSafely(v, "", "N8N_TEST_VAR")

	err := os.Setenv("N8N_VALID_VAR", "valid-value")
	assert.NoError(t, err)
	config.BindEnvSafely(v, "valid_var", "N8N_VALID_VAR")
	assert.Equal(t, "valid-value", v.GetString("valid_var"))
	err = os.Unsetenv("N8N_VALID_VAR")
	assert.NoError(t, err)
}
