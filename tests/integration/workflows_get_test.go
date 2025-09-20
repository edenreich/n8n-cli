// Package integration contains integration tests for the n8n-cli
package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupGetWorkflowTest creates a test server and configures the test environment
func setupGetWorkflowTest(t *testing.T, workflowID string, responseData string) (*httptest.Server, func()) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-N8N-API-KEY")
		if apiKey != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		expectedPath := fmt.Sprintf("/api/v1/workflows/%s", workflowID)
		if r.URL.Path == expectedPath {
			if workflowID != "123" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprint(w, `{"error": "Workflow not found"}`)
				return
			}

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

func TestGetWorkflow(t *testing.T) {
	tests := []struct {
		name          string
		workflowID    string
		responseData  string
		expectedError bool
		errorContains string
	}{
		{
			name:       "Successfully get workflow",
			workflowID: "123",
			responseData: `{
				"id": "123",
				"name": "Test Workflow 1",
				"active": true,
				"nodes": [
					{
						"id": "node1",
						"name": "Start",
						"type": "n8n-nodes-base.start"
					}
				],
				"connections": {}
			}`,
			expectedError: false,
		},
		{
			name:       "Workflow not found",
			workflowID: "999",
			responseData: `{
				"error": "Workflow not found"
			}`,
			expectedError: true,
			errorContains: "API returned error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server, cleanup := setupGetWorkflowTest(t, tc.workflowID, tc.responseData)
			defer cleanup()

			client := n8n.NewClient(server.URL, "test-api-key")
			workflow, err := client.GetWorkflow(tc.workflowID)

			if tc.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, workflow)
				assert.Equal(t, tc.workflowID, *workflow.Id)
				assert.NotNil(t, workflow.Name)
				assert.Equal(t, "Test Workflow 1", workflow.Name)
			}
		})
	}
}
