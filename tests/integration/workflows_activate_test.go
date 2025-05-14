// Package integration contains integration tests for the n8n-cli
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// setupActivateWorkflowTest creates a test server and configures the test environment
func setupActivateWorkflowTest(t *testing.T) (*httptest.Server, map[string]*n8n.Workflow, func()) {
	mockWorkflows := map[string]*n8n.Workflow{
		"123": {
			Id:   stringPtr("123"),
			Name: "Test Workflow 1",
		},
		"456": {
			Id:   stringPtr("456"),
			Name: "Test Workflow 2",
		},
		"nonexistent": nil,
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-N8N-API-KEY")
		if apiKey != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.Method == http.MethodPost {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) == 6 && parts[3] == "workflows" && parts[5] == "activate" {
				workflowID := parts[4]

				workflow, exists := mockWorkflows[workflowID]
				if !exists || workflow == nil {
					w.WriteHeader(http.StatusNotFound)
					_, _ = fmt.Fprintln(w, `{"error": "Workflow not found"}`)
					return
				}

				trueValue := true
				workflow.Active = &trueValue

				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(workflow)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprintln(w, `{"error": "Failed to encode response"}`)
					return
				}
				return
			}
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

	return mockServer, mockWorkflows, cleanup
}

// TestActivateWorkflow tests the activate workflow command
func TestActivateWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "Successfully activate existing workflow",
			args:           []string{"activate", "123"},
			expectedOutput: "Workflow with ID 123 has been activated successfully",
			expectedError:  false,
		},
		{
			name:           "Successfully activate another workflow",
			args:           []string{"activate", "456"},
			expectedOutput: "Workflow with ID 456 has been activated successfully",
			expectedError:  false,
		},
		{
			name:           "Attempt to activate non-existent workflow",
			args:           []string{"activate", "999"},
			expectedOutput: "Error activating workflow",
			expectedError:  true,
			errorContains:  "404",
		},
		{
			name:          "No workflow ID provided",
			args:          []string{"activate"},
			expectedError: true,
			errorContains: "accepts 1 arg(s), received 0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, cleanup := setupActivateWorkflowTest(t)
			defer cleanup()

			rootCmd := rootcmd.GetRootCmd()
			args := append([]string{"workflows"}, tc.args...)
			rootCmd.SetArgs(args)

			stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
			rootCmd.SetOut(stdout)
			rootCmd.SetErr(stderr)

			err := rootCmd.Execute()

			if tc.expectedError {
				assert.Error(t, err)
				errorOutput := stderr.String()
				if err != nil {
					errorOutput += err.Error()
				}
				assert.Contains(t, errorOutput, tc.errorContains)
			} else {
				assert.NoError(t, err, "Expected no error when executing activate command")
				assert.Contains(t, stdout.String(), tc.expectedOutput)
			}
		})
	}
}

// TestActivateWorkflowWithAuthError tests behavior when authentication fails
func TestActivateWorkflowWithAuthError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
	}))
	defer mockServer.Close()

	viper.Reset()
	viper.Set("api_key", "wrong-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()

	rootCmd := rootcmd.GetRootCmd()
	rootCmd.SetArgs([]string{"workflows", "activate", "123"})

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	err := rootCmd.Execute()

	assert.Error(t, err)
	errorOutput := stderr.String()
	if err != nil {
		errorOutput += err.Error()
	}
	assert.Contains(t, errorOutput, "401")
	assert.Contains(t, errorOutput, "Unauthorized")
}
