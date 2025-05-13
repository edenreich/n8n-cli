// Package config provides configuration functionality for the n8n-cli application
package config

// ConfigInterface defines the contract for configuration objects
//
//go:generate counterfeiter -o configfakes/fake_config.go . ConfigInterface
type ConfigInterface interface {
	// GetAPIToken returns the API token for authenticating with the n8n API
	GetAPIToken() string

	// GetAPIBaseURL returns the base URL for the n8n API
	GetAPIBaseURL() string
}

// Ensure Config implements ConfigInterface
var _ ConfigInterface = (*Config)(nil)
