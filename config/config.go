/*
Copyright © 2025 Eden Reich

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the global configuration for the n8n CLI
type Config struct {
	APIToken    string
	InstanceURL string
	APIBaseURL  string
}

// globalConfig holds the application configuration once loaded
var globalConfig *Config

// LoadConfig loads the configuration from environment variables
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	apiToken := os.Getenv("N8N_API_KEY")
	instanceURL := os.Getenv("N8N_INSTANCE_URL")

	if apiToken == "" || instanceURL == "" {
		return nil, errors.New("N8N_API_KEY and N8N_INSTANCE_URL environment variables must be set")
	}

	apiBaseURL := ensureAPIPrefix(instanceURL)

	config := &Config{
		APIToken:    apiToken,
		InstanceURL: instanceURL,
		APIBaseURL:  apiBaseURL,
	}

	globalConfig = config
	return config, nil
}

// GetConfig returns the global configuration or loads it if not already loaded
func GetConfig() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	return LoadConfig()
}

// ensureAPIPrefix ensures the URL has the /api/v1 prefix
func ensureAPIPrefix(url string) string {
	url = strings.TrimSuffix(url, "/")

	if !strings.HasSuffix(url, "/api/v1") {
		return url + "/api/v1"
	}

	return url
}

// InitConfig initializes the configuration during startup
func InitConfig() error {
	_, err := LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "You can create a .env file based on .env.example")
	}
	return err
}
