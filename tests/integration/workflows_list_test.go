// Package integration contains integration tests for the n8n-cli
package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/edenreich/n8n-cli/config/configfakes"
	"github.com/stretchr/testify/assert"
)

// TestListWorkflowsOutput tests that the list command outputs a table with ID, NAME, ACTIVE columns
func TestListWorkflowsOutput(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-N8N-API-KEY")
		if apiKey != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		// Return mock workflow data
		if r.URL.Path == "/api/v1/workflows" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{
				"data": [
					{
						"id": "123",
						"name": "Test Workflow 1",
						"active": true
					},
					{
						"id": "456",
						"name": "Test Workflow 2",
						"active": false
					},
					{
						"id": "789",
						"name": "Test Workflow 3",
						"active": true
					}
				],
				"nextCursor": null
			}`)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
	}))
	defer mockServer.Close()

	fakeConfig := &configfakes.FakeConfigInterface{}
	fakeConfig.GetAPITokenReturns("test-api-key")
	fakeConfig.GetAPIBaseURLReturns(mockServer.URL + "/api/v1")

	origGetConfigProvider := cmd.GetConfigProvider
	defer func() { cmd.GetConfigProvider = origGetConfigProvider }()
	cmd.GetConfigProvider = func() (cmd.ConfigProvider, error) {
		return fakeConfig, nil
	}

	stdout, stderr, err := executeCommand(t, workflows.ListCmd)

	assert.NoError(t, err, "Expected no error when executing list command")
	assert.Empty(t, stderr, "Expected no stderr output")

	assert.Contains(t, stdout, "ID")
	assert.Contains(t, stdout, "NAME")
	assert.Contains(t, stdout, "ACTIVE")

	assert.Contains(t, stdout, "123")
	assert.Contains(t, stdout, "Test Workflow 1")
	assert.Contains(t, stdout, "Test Workflow 2")
	assert.Contains(t, stdout, "Test Workflow 3")

	assert.True(t, strings.Contains(stdout, "true") || strings.Contains(stdout, "Yes") ||
		strings.Contains(stdout, "Active") || strings.Contains(stdout, "✓"),
		"Expected active status to be indicated for active workflows")
}
