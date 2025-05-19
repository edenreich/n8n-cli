// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func setupSyncWorkflowsTest(t *testing.T) (*httptest.Server, string, func()) {
	tmpDir, err := os.MkdirTemp("", "workflow-sync-test")
	require.NoError(t, err, "Failed to create temp directory")

	testWorkflows := map[string]struct {
		name   string
		id     string
		active bool
		format string
	}{
		"new_workflow.json":         {"New Workflow", "", false, "json"},
		"existing_workflow.json":    {"Existing Workflow", "456", false, "json"},
		"activate_workflow.json":    {"Activate Workflow", "789", true, "json"},
		"nonexistent_workflow.json": {"Nonexistent Workflow", "999", false, "json"},
		"yaml_workflow.yaml":        {"YAML Workflow", "", false, "yaml"},
	}

	for filename, info := range testWorkflows {
		workflow := createTestWorkflow(info.name, info.id)
		if info.active {
			*workflow.Active = true
		}

		switch info.format {
		case "json":
			writeWorkflowFile(t, tmpDir, filename, workflow)
		case "yaml":
			writeYAMLWorkflowFile(t, tmpDir, filename, workflow)
		}
	}

	existingWorkflows := []n8n.Workflow{
		func() n8n.Workflow {
			wf := createTestWorkflow("Existing Workflow", "456")
			return wf
		}(),
		func() n8n.Workflow {
			wf := createTestWorkflow("Activate Workflow", "789")
			return wf
		}(),
	}

	sendJSONResponse := func(w http.ResponseWriter, data interface{}) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintln(w, `{"error": "Failed to encode response"}`)
		}
	}

	sendErrorResponse := func(w http.ResponseWriter, statusCode int, message string) {
		w.WriteHeader(statusCode)
		_, _ = fmt.Fprintf(w, `{"error": "%s"}`, message)
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		switch {
		case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodGet:
			sendJSONResponse(w, struct {
				Data *[]n8n.Workflow `json:"data"`
			}{
				Data: &existingWorkflows,
			})

		case strings.HasPrefix(r.URL.Path, "/api/v1/workflows/") && r.Method == http.MethodGet &&
			!strings.HasSuffix(r.URL.Path, "/activate") && !strings.HasSuffix(r.URL.Path, "/deactivate"):
			id := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/")

			var foundWorkflow *n8n.Workflow
			for i := range existingWorkflows {
				if existingWorkflows[i].Id != nil && *existingWorkflows[i].Id == id {
					foundWorkflow = &existingWorkflows[i]
					break
				}
			}

			if foundWorkflow != nil {
				sendJSONResponse(w, foundWorkflow)
			} else {
				sendErrorResponse(w, http.StatusNotFound, "Workflow not found")
			}

		case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodPost:
			var wf n8n.Workflow
			if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
				sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
				return
			}

			newID := "123"
			wf.Id = &newID
			sendJSONResponse(w, wf)

		case strings.HasPrefix(r.URL.Path, "/api/v1/workflows/") && r.Method == http.MethodPut:
			id := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/")

			var wf n8n.Workflow
			if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
				sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
				return
			}

			wf.Id = &id
			sendJSONResponse(w, wf)

		case strings.HasSuffix(r.URL.Path, "/activate") && r.Method == http.MethodPost:
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/activate")
			active := true

			sendJSONResponse(w, n8n.Workflow{
				Id:     &id,
				Name:   "Activate Workflow",
				Active: &active,
			})

		case strings.HasSuffix(r.URL.Path, "/deactivate") && r.Method == http.MethodPost:
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/deactivate")
			active := false

			sendJSONResponse(w, n8n.Workflow{
				Id:     &id,
				Name:   "Deactivate Workflow",
				Active: &active,
			})

		default:
			sendErrorResponse(w, http.StatusNotFound, "Not found")
		}
	}))

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()

	cleanup := func() {
		err := os.RemoveAll(tmpDir)
		assert.NoError(t, err, "Failed to remove temp directory")
		mockServer.Close()
	}

	return mockServer, tmpDir, cleanup
}

// Helper to create a test workflow
func createTestWorkflow(name string, id string) n8n.Workflow {
	active := false
	var workflowID *string
	if id != "" {
		workflowID = &id
	}

	return n8n.Workflow{
		Id:          workflowID,
		Name:        name,
		Active:      &active,
		Nodes:       []n8n.Node{},
		Connections: map[string]interface{}{},
		Settings:    n8n.WorkflowSettings{},
	}
}

// Helper to write a workflow to a JSON file
func writeWorkflowFile(t *testing.T, dir string, filename string, workflow n8n.Workflow) {
	data, err := json.MarshalIndent(workflow, "", "  ")
	require.NoError(t, err, "Failed to marshal workflow")

	err = os.WriteFile(filepath.Join(dir, filename), data, 0644)
	require.NoError(t, err, "Failed to write workflow file")
}

// Helper to write a workflow to a YAML file
func writeYAMLWorkflowFile(t *testing.T, dir string, filename string, workflow n8n.Workflow) {
	workflowMap := map[string]interface{}{
		"name":        workflow.Name,
		"nodes":       workflow.Nodes,
		"connections": workflow.Connections,
		"settings":    workflow.Settings,
	}

	if workflow.Id != nil {
		workflowMap["id"] = *workflow.Id
	}

	if workflow.Active != nil {
		workflowMap["active"] = *workflow.Active
	}

	data, err := yaml.Marshal(workflowMap)
	require.NoError(t, err, "Failed to marshal workflow to YAML")

	err = os.WriteFile(filepath.Join(dir, filename), data, 0644)
	require.NoError(t, err, "Failed to write YAML workflow file")
}

func TestSyncWorkflows(t *testing.T) {
	_, tmpDir, cleanup := setupSyncWorkflowsTest(t)
	defer cleanup()

	pruneTestDir, err := os.MkdirTemp("", "workflow-prune-test")
	require.NoError(t, err, "Failed to create temp directory for prune tests")
	defer func() {
		err := os.RemoveAll(pruneTestDir)
		assert.NoError(t, err, "Failed to remove temp directory for prune tests")
	}()

	createWorkflowFile(t, pruneTestDir, "workflow1.json", "1", "Workflow 1", false)
	createWorkflowFile(t, pruneTestDir, "workflow2.json", "2", "Workflow 2", false)

	tests := []struct {
		name             string
		args             []string
		directory        string
		expectedError    bool
		setupMockServer  func() (*httptest.Server, []string, func())
		validateOutput   func(t *testing.T, stdout string)
		validateRequests func(t *testing.T, requests []string)
	}{
		{
			name:          "Sync JSON workflows",
			args:          []string{"--directory", tmpDir},
			directory:     tmpDir,
			expectedError: false,
			setupMockServer: func() (*httptest.Server, []string, func()) {
				var requests []string
				server := setupBasicMockServer(&requests)
				return server, requests, func() { server.Close() }
			},
			validateOutput: func(t *testing.T, stdout string) {
				t.Logf("Command output: %s", stdout)
				assert.Contains(t, stdout, "Created workflow")
			},
			validateRequests: func(t *testing.T, requests []string) {},
		},
		{
			name:          "Dry run",
			args:          []string{"--directory", tmpDir, "--dry-run"},
			directory:     tmpDir,
			expectedError: false,
			setupMockServer: func() (*httptest.Server, []string, func()) {
				var requests []string
				server := setupBasicMockServer(&requests)
				return server, requests, func() { server.Close() }
			},
			validateOutput: func(t *testing.T, stdout string) {
				assert.Contains(t, stdout, "Would create workflow")
				assert.Contains(t, stdout, "Would create workflow")
				assert.NotContains(t, stdout, "Created workflow")
				assert.NotContains(t, stdout, "Updated workflow")
			},
			validateRequests: func(t *testing.T, requests []string) {},
		},
		{
			name:          "Invalid directory",
			args:          []string{"--directory", "/nonexistent-dir"},
			directory:     "",
			expectedError: true,
			setupMockServer: func() (*httptest.Server, []string, func()) {
				var requests []string
				server := setupBasicMockServer(&requests)
				return server, requests, func() { server.Close() }
			},
			validateOutput:   func(t *testing.T, stdout string) {},
			validateRequests: func(t *testing.T, requests []string) {},
		}, {
			name:          "Prune workflows",
			args:          []string{"--directory", pruneTestDir, "--prune"},
			directory:     pruneTestDir,
			expectedError: false,
			setupMockServer: func() (*httptest.Server, []string, func()) {
				var requests []string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestPath := r.Method + " " + r.URL.Path
					requests = append(requests, requestPath)

					apiKey := r.Header.Get("X-N8N-API-KEY")
					if apiKey != "test-api-key" {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
						return
					}

					switch {
					case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodGet:
						w.Header().Set("Content-Type", "application/json")
						_, _ = fmt.Fprintln(w, `{
							"data": [
								{"id": "1", "name": "Workflow 1", "active": false},
								{"id": "2", "name": "Workflow 2", "active": false},
								{"id": "3", "name": "Workflow 3", "active": false}
							]
						}`)
						return

					case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodPost:
						var wf n8n.Workflow
						err := json.NewDecoder(r.Body).Decode(&wf)
						if err != nil {
							w.WriteHeader(http.StatusBadRequest)
							_, _ = fmt.Fprintln(w, `{"error": "Invalid request body"}`)
							return
						}

						newID := "new-id"
						wf.Id = &newID

						w.Header().Set("Content-Type", "application/json")
						err = json.NewEncoder(w).Encode(wf)
						if err != nil {
							w.WriteHeader(http.StatusInternalServerError)
							_, _ = fmt.Fprintln(w, `{"error": "Failed to encode response"}`)
						}
						return

					case strings.HasPrefix(r.URL.Path, "/api/v1/workflows/") && r.Method == http.MethodPut:
						parts := strings.Split(r.URL.Path, "/")
						if len(parts) == 5 {
							workflowID := parts[4]
							w.Header().Set("Content-Type", "application/json")
							_, _ = fmt.Fprintf(w, `{"id": "%s", "name": "Workflow %s"}`, workflowID, workflowID)
							return
						}

					case strings.HasSuffix(r.URL.Path, "/activate") && r.Method == http.MethodPost:
						id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/activate")
						w.Header().Set("Content-Type", "application/json")
						_, _ = fmt.Fprintf(w, `{"id": "%s", "name": "Workflow %s", "active": true}`, id, id)
						return

					case r.URL.Path == "/api/v1/workflows/3" && r.Method == http.MethodDelete:
						w.WriteHeader(http.StatusOK)
						return

					case strings.HasPrefix(r.URL.Path, "/api/v1/workflows/") && r.Method == http.MethodDelete:
						parts := strings.Split(r.URL.Path, "/")
						if len(parts) == 5 {
							w.WriteHeader(http.StatusOK)
							return
						}
					}

					w.WriteHeader(http.StatusNotFound)
					_, _ = fmt.Fprintln(w, `{"error": "Not found"}`)
				}))
				return server, requests, func() { server.Close() }
			},
			validateOutput: func(t *testing.T, stdout string) {
				t.Logf("Command output: %s", stdout)
				hasDeleted := strings.Contains(stdout, "Deleted workflow 'Workflow 3' (ID: 3)")
				hasWouldDelete := strings.Contains(stdout, "Would delete workflow 'Workflow 3' (ID: 3)")
				if !hasDeleted && !hasWouldDelete {
					assert.Fail(t, "Expected either 'Deleted workflow' or 'Would delete workflow' in output")
				}
			},
			validateRequests: func(t *testing.T, requests []string) {
				t.Logf("Requests: %v", requests)
			},
		},
		{
			name:          "Dry run with prune",
			args:          []string{"--directory", pruneTestDir, "--prune", "--dry-run"},
			directory:     pruneTestDir,
			expectedError: false,
			setupMockServer: func() (*httptest.Server, []string, func()) {
				var requests []string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requestPath := r.Method + " " + r.URL.Path
					requests = append(requests, requestPath)

					apiKey := r.Header.Get("X-N8N-API-KEY")
					if apiKey != "test-api-key" {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
						return
					}

					if r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodGet {
						w.Header().Set("Content-Type", "application/json")
						_, _ = fmt.Fprintln(w, `{
							"data": [
								{"id": "1", "name": "Workflow 1", "active": false},
								{"id": "2", "name": "Workflow 2", "active": false},
								{"id": "3", "name": "Workflow 3", "active": false}
							]
						}`)
						return
					}

					w.WriteHeader(http.StatusNotFound)
					_, _ = fmt.Fprintln(w, `{"error": "Not found"}`)
				}))
				return server, requests, func() { server.Close() }
			},
			validateOutput: func(t *testing.T, stdout string) {
				t.Logf("Command output: %s", stdout)
				assert.Contains(t, stdout, "Would delete workflow 'Workflow 3' (ID: 3)")
			},
			validateRequests: func(t *testing.T, requests []string) {
				t.Logf("Requests: %v", requests)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server, requests, serverCleanup := tc.setupMockServer()
			defer serverCleanup()

			viper.Reset()
			viper.Set("api_key", "test-api-key")
			viper.Set("instance_url", server.URL)
			config.Initialize()

			stdout, stderr, err := executeCommand(t, workflows.SyncCmd, tc.args...)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "Expected no error when executing sync command")
				assert.Empty(t, stderr, "Expected no stderr output")
				tc.validateOutput(t, stdout)
				tc.validateRequests(t, requests)
			}
		})
	}
}

// Helper function for setting up a basic mock server
func setupBasicMockServer(requests *[]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-N8N-API-KEY")
		if apiKey != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		*requests = append(*requests, r.Method+" "+r.URL.Path)

		switch {
		case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodGet:
			// Create workflows using the same helper function for consistency
			existingWorkflows := []n8n.Workflow{
				createTestWorkflow("Existing Workflow", "456"),
				createTestWorkflow("Activate Workflow", "789"),
			}

			// Format workflows list response
			workflowsResponse := struct {
				Data *[]n8n.Workflow `json:"data"`
			}{
				Data: &existingWorkflows,
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(workflowsResponse)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintln(w, `{"error": "Failed to encode workflows list"}`)
			}
			return
		case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodPost:
			var wf n8n.Workflow
			err := json.NewDecoder(r.Body).Decode(&wf)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprintln(w, `{"error": "Invalid request body"}`)
				return
			}

			newID := "new-id"
			wf.Id = &newID

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(wf)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintln(w, `{"error": "Failed to encode response"}`)
			}
			return

		case strings.HasPrefix(r.URL.Path, "/api/v1/workflows/") && r.Method == http.MethodPut:
			id := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/")

			var wf n8n.Workflow
			err := json.NewDecoder(r.Body).Decode(&wf)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprintln(w, `{"error": "Invalid request body"}`)
				return
			}

			wf.Id = &id

			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(wf)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintln(w, `{"error": "Failed to encode response"}`)
			}
			return

		case strings.HasSuffix(r.URL.Path, "/activate") && r.Method == http.MethodPost:
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/activate")

			active := true
			resp := n8n.Workflow{
				Id:     &id,
				Name:   "Activate Workflow",
				Active: &active,
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintln(w, `{"error": "Failed to encode response"}`)
			}
			return

		case strings.HasSuffix(r.URL.Path, "/deactivate") && r.Method == http.MethodPost:
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/deactivate")

			active := false
			resp := n8n.Workflow{
				Id:     &id,
				Name:   id,
				Active: &active,
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = fmt.Fprintln(w, `{"error": "Failed to encode response"}`)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintln(w, `{"error": "Not found"}`)
	}))
}

// Helper function to create a test workflow file
func createWorkflowFile(t *testing.T, dir, filename, id, name string, active bool) {
	workflowData := map[string]interface{}{
		"id":          id,
		"name":        name,
		"active":      active,
		"nodes":       []n8n.Node{},
		"connections": map[string]interface{}{},
		"settings":    map[string]interface{}{},
	}

	data, err := json.MarshalIndent(workflowData, "", "  ")
	require.NoError(t, err, "Failed to marshal workflow")

	err = os.WriteFile(filepath.Join(dir, filename), data, 0644)
	require.NoError(t, err, "Failed to write workflow file")
}
