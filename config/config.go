// Package config provides configuration functionality for the n8n-cli application
package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	APIToken   string
	APIBaseURL string
}

// GetAPIToken returns the API token
func (c *Config) GetAPIToken() string {
	return c.APIToken
}

// GetAPIBaseURL returns the API base URL
func (c *Config) GetAPIBaseURL() string {
	return c.APIBaseURL
}

// GetConfig returns the application configuration using the global viper instance
func GetConfig() (*Config, error) {
	return LoadConfig(viper.GetViper())
}

// LoadConfig creates a Config from a Viper instance
func LoadConfig(v *viper.Viper) (*Config, error) {
	apiToken := v.GetString("api_key")
	if apiToken == "" {
		return nil, fmt.Errorf("N8N API key is not set. Use --api-key flag or set N8N_API_KEY environment variable")
	}

	instanceURL := v.GetString("instance_url")
	if instanceURL == "" {
		return nil, fmt.Errorf("N8N instance URL is not set. Use --url flag or set N8N_INSTANCE_URL environment variable")
	}

	return &Config{
		APIToken:   apiToken,
		APIBaseURL: strings.TrimSuffix(instanceURL, "/") + "/api/v1",
	}, nil
}

// FileReader is an interface for reading files
type FileReader interface {
	Open(name string) (io.ReadCloser, error)
}

// OSFileReader implements FileReader using os package
type OSFileReader struct{}

// Open opens a file using os.Open
func (r *OSFileReader) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// DefaultFileReader is the default file reader
var DefaultFileReader FileReader = &OSFileReader{}

// LoadEnvFile loads environment variables from a .env file if it exists
func LoadEnvFile() {
	LoadEnvFileWithReader(DefaultFileReader, viper.GetViper())
}

// LoadEnvFileWithReader loads environment variables from a .env file using the provided reader
func LoadEnvFileWithReader(reader FileReader, v *viper.Viper) {
	envFile, err := reader.Open(".env")
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
				v.Set(viperKey, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error reading .env file: %v\n", err)
	}
}

// InitConfig sets up the configuration system
func InitConfig() {
	v := viper.GetViper()
	v.SetEnvPrefix("N8N")
	v.AutomaticEnv()

	// Bind environment variables
	_ = v.BindEnv("api_key", "N8N_API_KEY")
	_ = v.BindEnv("instance_url", "N8N_INSTANCE_URL")

	// Set defaults
	v.SetDefault("instance_url", "http://localhost:5678")
	v.SetDefault("api_key", "")
}
