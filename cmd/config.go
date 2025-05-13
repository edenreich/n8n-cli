// Package cmd contains commands for the n8n-cli
package cmd

import (
	"github.com/edenreich/n8n-cli/config"
)

// ConfigProvider is an interface for configuration providers
type ConfigProvider interface {
	GetAPIToken() string
	GetAPIBaseURL() string
}

// GetConfigProvider is a function that returns a configuration provider
// This can be overridden in tests to provide a mock configuration
var GetConfigProvider = func() (ConfigProvider, error) {
	return config.GetConfig()
}
