// Package config provides configuration functionality for the n8n-cli application
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	APIToken   string
	APIBaseURL string
}

// GetConfig returns the application configuration using Viper
func GetConfig() (*Config, error) {
	apiToken := viper.GetString("api_key")
	if apiToken == "" {
		return nil, fmt.Errorf("N8N API key is not set. Use --api-key flag or set N8N_API_KEY environment variable")
	}

	instanceURL := viper.GetString("instance_url")
	if instanceURL == "" {
		return nil, fmt.Errorf("N8N instance URL is not set. Use --url flag or set N8N_INSTANCE_URL environment variable")
	}

	return &Config{
		APIToken:   apiToken,
		APIBaseURL: strings.TrimSuffix(instanceURL, "/") + "/api/v1",
	}, nil
}

// LoadEnvFile loads environment variables from a .env file if it exists
func LoadEnvFile() {
	envFile, err := os.Open(".env")
	if err != nil {
		return
	}
	defer func() {
		if err := envFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error closing .env file: %v\n", err)
		}
	}()
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		if strings.HasPrefix(key, "N8N_") {
			viperKey := strings.ToLower(strings.TrimPrefix(key, "N8N_"))
			if os.Getenv(key) == "" {
				viper.Set(viperKey, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error reading .env file: %v\n", err)
	}
}
