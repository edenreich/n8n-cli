package config

import (
	"os"
	"testing"
)

func TestEnsureAPIPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL without trailing slash",
			input:    "https://n8n.example.com",
			expected: "https://n8n.example.com/api/v1",
		},
		{
			name:     "URL with trailing slash",
			input:    "https://n8n.example.com/",
			expected: "https://n8n.example.com/api/v1",
		},
		{
			name:     "URL already with API prefix without trailing slash",
			input:    "https://n8n.example.com/api/v1",
			expected: "https://n8n.example.com/api/v1",
		},
		{
			name:     "URL already with API prefix with trailing slash",
			input:    "https://n8n.example.com/api/v1/",
			expected: "https://n8n.example.com/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureAPIPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("ensureAPIPrefix(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	originalAPIKey := os.Getenv("N8N_API_KEY")
	originalInstanceURL := os.Getenv("N8N_INSTANCE_URL")
	defer func() {
		os.Setenv("N8N_API_KEY", originalAPIKey)
		os.Setenv("N8N_INSTANCE_URL", originalInstanceURL)
		globalConfig = nil
	}()

	t.Run("Both environment variables set", func(t *testing.T) {
		os.Setenv("N8N_API_KEY", "test-api-key")
		os.Setenv("N8N_INSTANCE_URL", "https://n8n.example.com")
		globalConfig = nil

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v, want nil", err)
		}
		if config == nil {
			t.Fatal("LoadConfig() returned nil config")
		}
		if config.APIToken != "test-api-key" {
			t.Errorf("config.APIToken = %q, want %q", config.APIToken, "test-api-key")
		}
		if config.InstanceURL != "https://n8n.example.com" {
			t.Errorf("config.InstanceURL = %q, want %q", config.InstanceURL, "https://n8n.example.com")
		}
		if config.APIBaseURL != "https://n8n.example.com/api/v1" {
			t.Errorf("config.APIBaseURL = %q, want %q", config.APIBaseURL, "https://n8n.example.com/api/v1")
		}
	})

	t.Run("Missing API key", func(t *testing.T) {
		os.Setenv("N8N_API_KEY", "")
		os.Setenv("N8N_INSTANCE_URL", "https://n8n.example.com")
		globalConfig = nil

		_, err := LoadConfig()
		if err == nil {
			t.Error("LoadConfig() error = nil, want error")
		}
	})

	t.Run("Missing instance URL", func(t *testing.T) {
		os.Setenv("N8N_API_KEY", "test-api-key")
		os.Setenv("N8N_INSTANCE_URL", "")
		globalConfig = nil

		_, err := LoadConfig()
		if err == nil {
			t.Error("LoadConfig() error = nil, want error")
		}
	})
}

func TestGetConfig(t *testing.T) {
	originalAPIKey := os.Getenv("N8N_API_KEY")
	originalInstanceURL := os.Getenv("N8N_INSTANCE_URL")
	defer func() {
		os.Setenv("N8N_API_KEY", originalAPIKey)
		os.Setenv("N8N_INSTANCE_URL", originalInstanceURL)
		globalConfig = nil
	}()

	t.Run("Config already loaded", func(t *testing.T) {
		os.Setenv("N8N_API_KEY", "test-api-key")
		os.Setenv("N8N_INSTANCE_URL", "https://n8n.example.com")
		globalConfig = nil

		firstConfig, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v, want nil", err)
		}

		secondConfig, err := GetConfig()
		if err != nil {
			t.Fatalf("GetConfig() error = %v, want nil", err)
		}

		if secondConfig != firstConfig {
			t.Error("GetConfig() returned different config instance than LoadConfig()")
		}
	})

	t.Run("Config not loaded yet", func(t *testing.T) {
		globalConfig = nil
		os.Setenv("N8N_API_KEY", "test-api-key")
		os.Setenv("N8N_INSTANCE_URL", "https://n8n.example.com")

		config, err := GetConfig()
		if err != nil {
			t.Fatalf("GetConfig() error = %v, want nil", err)
		}
		if config == nil {
			t.Fatal("GetConfig() returned nil config")
		}
		if config.APIToken != "test-api-key" {
			t.Errorf("config.APIToken = %q, want %q", config.APIToken, "test-api-key")
		}
	})
}

func TestInitConfig(t *testing.T) {
	originalAPIKey := os.Getenv("N8N_API_KEY")
	originalInstanceURL := os.Getenv("N8N_INSTANCE_URL")
	defer func() {
		os.Setenv("N8N_API_KEY", originalAPIKey)
		os.Setenv("N8N_INSTANCE_URL", originalInstanceURL)
		globalConfig = nil
	}()

	t.Run("Success initialization", func(t *testing.T) {
		os.Setenv("N8N_API_KEY", "test-api-key")
		os.Setenv("N8N_INSTANCE_URL", "https://n8n.example.com")
		globalConfig = nil

		err := InitConfig()
		if err != nil {
			t.Fatalf("InitConfig() error = %v, want nil", err)
		}

		if globalConfig == nil {
			t.Error("InitConfig() did not set globalConfig")
		}
	})

	t.Run("Error initialization", func(t *testing.T) {
		os.Setenv("N8N_API_KEY", "")
		os.Setenv("N8N_INSTANCE_URL", "")
		globalConfig = nil

		err := InitConfig()
		if err == nil {
			t.Error("InitConfig() error = nil, want error")
		}
	})
}
