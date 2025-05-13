package tests

import (
	"testing"

	"github.com/edenreich/n8n-cli/config/configfakes"
)

// NewMockConfig returns a mock config interface for testing
func NewMockConfig() *configfakes.FakeConfigInterface {
	fakeConfig := &configfakes.FakeConfigInterface{}
	fakeConfig.GetAPITokenReturns("test-api-token")
	fakeConfig.GetAPIBaseURLReturns("http://test-api:5678/api/v1")
	return fakeConfig
}

// SkipIfNotIntegration skips the test if environment variable INTEGRATION_TESTS is not set
func SkipIfNotIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}
