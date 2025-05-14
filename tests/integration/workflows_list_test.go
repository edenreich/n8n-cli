// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// setupListWorkflowsTest creates a test server and configures the test environment
func setupListWorkflowsTest(t *testing.T, responseData string) (*httptest.Server, func()) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-N8N-API-KEY")
		if apiKey != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/workflows" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, responseData)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
	}))

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()

	cleanup := func() {
		mockServer.Close()
	}

	return mockServer, cleanup
}

func TestListWorkflows(t *testing.T) {
	mockedResponse := `{
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
	}`

	tests := []struct {
		name           string
		outputFormat   string
		responseData   string
		expectedError  bool
		errorContains  string
		validateOutput func(t *testing.T, stdout string)
	}{
		{
			name:          "Table output format",
			outputFormat:  "table",
			responseData:  mockedResponse,
			expectedError: false,
			validateOutput: func(t *testing.T, stdout string) {
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
			},
		},
		{
			name:          "JSON output format",
			outputFormat:  "json",
			responseData:  mockedResponse,
			expectedError: false,
			validateOutput: func(t *testing.T, stdout string) {
				var parsedWorkflows []n8n.Workflow
				err := json.Unmarshal([]byte(stdout), &parsedWorkflows)
				require.NoError(t, err, "Expected to parse valid JSON")

				assert.Equal(t, 3, len(parsedWorkflows), "Expected 3 workflows")
				assert.Equal(t, "123", *parsedWorkflows[0].Id, "Expected ID to match")
				assert.Equal(t, "Test Workflow 1", parsedWorkflows[0].Name, "Expected name to match")
				assert.True(t, *parsedWorkflows[0].Active, "Expected first workflow to be active")
				assert.Equal(t, "456", *parsedWorkflows[1].Id, "Expected ID to match")
				assert.Equal(t, "Test Workflow 2", parsedWorkflows[1].Name, "Expected name to match")
				assert.False(t, *parsedWorkflows[1].Active, "Expected second workflow to be inactive")
				assert.Equal(t, "789", *parsedWorkflows[2].Id, "Expected ID to match")
				assert.Equal(t, "Test Workflow 3", parsedWorkflows[2].Name, "Expected name to match")
				assert.True(t, *parsedWorkflows[2].Active, "Expected third workflow to be active")
			},
		},
		{
			name:          "YAML output format",
			outputFormat:  "yaml",
			responseData:  mockedResponse,
			expectedError: false,
			validateOutput: func(t *testing.T, stdout string) {
				var parsedWorkflows []n8n.Workflow
				err := yaml.Unmarshal([]byte(stdout), &parsedWorkflows)
				require.NoError(t, err, "Expected to parse valid YAML")

				assert.Equal(t, 3, len(parsedWorkflows), "Expected 3 workflows")
				assert.Equal(t, "123", *parsedWorkflows[0].Id, "Expected ID to match")
				assert.Equal(t, "Test Workflow 1", parsedWorkflows[0].Name, "Expected name to match")
				assert.True(t, *parsedWorkflows[0].Active, "Expected first workflow to be active")
				assert.Equal(t, "456", *parsedWorkflows[1].Id, "Expected ID to match")
				assert.Equal(t, "Test Workflow 2", parsedWorkflows[1].Name, "Expected name to match")
				assert.False(t, *parsedWorkflows[1].Active, "Expected second workflow to be inactive")
				assert.Equal(t, "789", *parsedWorkflows[2].Id, "Expected ID to match")
				assert.Equal(t, "Test Workflow 3", parsedWorkflows[2].Name, "Expected name to match")
				assert.True(t, *parsedWorkflows[2].Active, "Expected third workflow to be active")
			},
		},
		{
			name:           "Invalid output format",
			outputFormat:   "xml",
			responseData:   mockedResponse,
			expectedError:  true,
			errorContains:  "unsupported output format: xml",
			validateOutput: func(t *testing.T, stdout string) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, cleanup := setupListWorkflowsTest(t, tc.responseData)
			defer cleanup()

			if tc.outputFormat != "table" {
				err := workflows.ListCmd.Flags().Set("output", tc.outputFormat)
				require.NoError(t, err, "Expected to set output format flag")
			}

			defer func() {
				err := workflows.ListCmd.Flags().Set("output", "table")
				require.NoError(t, err, "Expected to reset output format flag")
			}()

			stdout, stderr, err := executeCommand(t, workflows.ListCmd)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err, "Expected no error when executing list command")
				assert.Empty(t, stderr, "Expected no stderr output")
				tc.validateOutput(t, stdout)
			}
		})
	}
}
