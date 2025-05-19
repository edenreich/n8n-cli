// filepath: /workspaces/n8n-cli/tests/integration/workflows_sync_refresh_test.go
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
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncWorkflows_Refresh(t *testing.T) {
	testCases := []struct {
		name           string
		refreshFlag    bool
		expectIDInFile bool
		args           []string
	}{
		{
			name:           "Default behavior (refresh=true)",
			refreshFlag:    true,
			expectIDInFile: true,
			args:           []string{"--directory"},
		},
		{
			name:           "Explicit refresh=false",
			refreshFlag:    false,
			expectIDInFile: false,
			args:           []string{"--directory", "", "--refresh=false"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "workflow-sync-refresh-test")
			require.NoError(t, err)
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Fatalf("Failed to remove temp directory: %v", err)
				}
			}()

			noIDWorkflow := n8n.Workflow{
				Name:        "New Workflow",
				Nodes:       []n8n.Node{},
				Connections: map[string]interface{}{},
				Settings:    n8n.WorkflowSettings{},
			}
			inactive := false
			noIDWorkflow.Active = &inactive

			noIDWorkflowBytes, err := json.MarshalIndent(noIDWorkflow, "", "  ")
			require.NoError(t, err)

			noIDFilePath := filepath.Join(tmpDir, "new_workflow.json")
			require.NoError(t, os.WriteFile(noIDFilePath, noIDWorkflowBytes, 0644))

			nonexistentID := "nonexistent-id"
			nonExistentWorkflow := n8n.Workflow{
				Id:          &nonexistentID,
				Name:        "Nonexistent ID Workflow",
				Nodes:       []n8n.Node{},
				Connections: map[string]interface{}{},
				Settings:    n8n.WorkflowSettings{},
			}
			nonExistentWorkflow.Active = &inactive

			nonExistentWorkflowBytes, err := json.MarshalIndent(nonExistentWorkflow, "", "  ")
			require.NoError(t, err)

			nonexistentIDFilePath := filepath.Join(tmpDir, "nonexistent_id_workflow.json")
			require.NoError(t, os.WriteFile(nonexistentIDFilePath, nonExistentWorkflowBytes, 0644))

			var requestCounter int
			var createdWorkflows []n8n.Workflow

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("Request %d: %s %s\n", requestCounter, r.Method, r.URL.Path)
				requestCounter++

				if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
					return
				}

				switch {
				case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodGet:
					w.Header().Set("Content-Type", "application/json")
					response := n8n.WorkflowList{
						Data: &createdWorkflows,
					}
					_ = json.NewEncoder(w).Encode(response)

				case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodPost:
					var workflow n8n.Workflow
					err := json.NewDecoder(r.Body).Decode(&workflow)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = fmt.Fprintln(w, `{"error": "Invalid workflow data"}`)
						return
					}

					newID := fmt.Sprintf("generated-id-%d", len(createdWorkflows)+1)
					workflow.Id = &newID

					createdWorkflows = append(createdWorkflows, workflow)

					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(workflow)

				case strings.HasSuffix(r.URL.Path, "/activate") && r.Method == http.MethodPost:

					id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/activate")

					var foundWorkflow *n8n.Workflow
					for i := range createdWorkflows {
						if createdWorkflows[i].Id != nil && *createdWorkflows[i].Id == id {
							foundWorkflow = &createdWorkflows[i]
							active := true
							foundWorkflow.Active = &active
							break
						}
					}

					if foundWorkflow != nil {
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(foundWorkflow)
					} else {
						w.WriteHeader(http.StatusNotFound)
						_, _ = fmt.Fprintf(w, `{"error": "Workflow with ID %s not found"}`, id)
					}
					return

				case strings.HasSuffix(r.URL.Path, "/deactivate") && r.Method == http.MethodPost:
					id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/deactivate")

					var foundWorkflow *n8n.Workflow
					for i := range createdWorkflows {
						if createdWorkflows[i].Id != nil && *createdWorkflows[i].Id == id {
							foundWorkflow = &createdWorkflows[i]
							inactive := false
							foundWorkflow.Active = &inactive
							break
						}
					}

					if foundWorkflow != nil {
						w.Header().Set("Content-Type", "application/json")
						_ = json.NewEncoder(w).Encode(foundWorkflow)
					} else {
						w.WriteHeader(http.StatusNotFound)
						_, _ = fmt.Fprintf(w, `{"error": "Workflow with ID %s not found"}`, id)
					}
					return

				case strings.HasPrefix(r.URL.Path, "/api/v1/workflows/") && r.Method == http.MethodGet:
					id := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/")

					for i := range createdWorkflows {
						if createdWorkflows[i].Id != nil && *createdWorkflows[i].Id == id {
							w.Header().Set("Content-Type", "application/json")
							_ = json.NewEncoder(w).Encode(createdWorkflows[i])
							return
						}
					}

					w.WriteHeader(http.StatusNotFound)
					_, _ = fmt.Fprintf(w, `{"error": "Workflow with ID %s not found"}`, id)

				case strings.HasPrefix(r.URL.Path, "/api/v1/workflows/") && r.Method == http.MethodPut:
					id := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/")

					var workflow n8n.Workflow
					err := json.NewDecoder(r.Body).Decode(&workflow)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = fmt.Fprintln(w, `{"error": "Invalid workflow data"}`)
						return
					}

					found := false
					for i := range createdWorkflows {
						if createdWorkflows[i].Id != nil && *createdWorkflows[i].Id == id {
							workflow.Id = &id
							createdWorkflows[i] = workflow
							found = true
							break
						}
					}

					if !found {
						w.WriteHeader(http.StatusNotFound)
						_, _ = fmt.Fprintf(w, `{"error": "Workflow with ID %s not found"}`, id)
						return
					}

					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(workflow)
				}
			}))
			defer server.Close()

			viper.Set("api_key", "test-api-key")
			viper.Set("instance_url", server.URL)
			defer viper.Reset()

			var args []string
			if tc.refreshFlag {
				args = []string{"--directory", tmpDir}
			} else {
				args = []string{"--directory", tmpDir, "--refresh=false"}
			}

			stdout, stderr, err := executeCommand(t, workflows.SyncCmd, args...)
			assert.NoError(t, err, "Command should succeed: %s\nstderr: %s", stdout, stderr)

			if !tc.expectIDInFile {
				content, err := os.ReadFile(noIDFilePath)
				require.NoError(t, err)

				var workflow n8n.Workflow
				err = json.Unmarshal(content, &workflow)
				require.NoError(t, err)

				assert.Nil(t, workflow.Id, "With refresh=false, the file should not have been updated with an ID")

				content, err = os.ReadFile(nonexistentIDFilePath)
				require.NoError(t, err)

				workflow = n8n.Workflow{}
				err = json.Unmarshal(content, &workflow)
				require.NoError(t, err)

				assert.NotNil(t, workflow.Id, "ID field should exist")
				assert.Equal(t, "nonexistent-id", *workflow.Id, "With refresh=false, the file should retain its original ID")
			} else {
				files, err := os.ReadDir(tmpDir)
				require.NoError(t, err)

				idUpdated := false

				for _, file := range files {
					if file.IsDir() {
						continue
					}

					filePath := filepath.Join(tmpDir, file.Name())
					content, err := os.ReadFile(filePath)
					require.NoError(t, err)

					var workflow n8n.Workflow
					err = json.Unmarshal(content, &workflow)
					require.NoError(t, err)

					if workflow.Id != nil && strings.HasPrefix(*workflow.Id, "generated-id-") {
						idUpdated = true
						break
					}
				}

				assert.True(t, idUpdated, "Expected at least one workflow file to be updated with a new ID")
			}
		})
	}
}
