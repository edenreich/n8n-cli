// Package unit contains unit tests for the n8n-cli
package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/config/configfakes"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportWorkflowByIDWithConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "n8n-cli-test-*")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	testCases := []struct {
		name           string
		workflowID     string
		responseStatus int
		responseBody   interface{}
		dryRun         bool
		expectedErr    bool
		expectedFile   string
	}{
		{
			name:           "Success - Import workflow",
			workflowID:     "123",
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"data": n8n.Workflow{
					Id:          stringPtr("123"),
					Name:        "Test Workflow",
					Active:      boolPtr(true),
					Nodes:       []n8n.Node{},
					Connections: map[string]interface{}{},
					Settings:    n8n.WorkflowSettings{},
				},
			},
			dryRun:       false,
			expectedErr:  false,
			expectedFile: "Test_Workflow.json",
		},
		{
			name:           "DryRun - Don't create file",
			workflowID:     "456",
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"data": n8n.Workflow{
					Id:          stringPtr("456"),
					Name:        "Another Workflow",
					Active:      boolPtr(false),
					Nodes:       []n8n.Node{},
					Connections: map[string]interface{}{},
					Settings:    n8n.WorkflowSettings{},
				},
			},
			dryRun:       true,
			expectedErr:  false,
			expectedFile: "Another_Workflow.json",
		},
		{
			name:           "Error - API error",
			workflowID:     "789",
			responseStatus: http.StatusNotFound,
			responseBody:   map[string]interface{}{"error": "Workflow not found"},
			dryRun:         false,
			expectedErr:    true,
			expectedFile:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/workflows/"+tc.workflowID, r.URL.Path)

				assert.Equal(t, "test-api-key", r.Header.Get("X-N8N-API-KEY"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(tc.responseStatus)

				respBytes, _ := json.Marshal(tc.responseBody)
				_, _ = w.Write(respBytes)
			}))
			defer mockServer.Close()

			fakeConfig := &configfakes.FakeConfigInterface{}
			fakeConfig.GetAPITokenReturns("test-api-key")
			fakeConfig.GetAPIBaseURLReturns(cmd.FormatAPIBaseURL(mockServer.URL))

			err := cmd.ImportWorkflowByIDWithConfig(
				fakeConfig,
				tempDir,
				tc.workflowID,
				tc.dryRun,
				false,
			)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				expectedFilePath := filepath.Join(tempDir, tc.expectedFile)
				_, fileErr := os.Stat(expectedFilePath)

				if tc.dryRun {
					assert.True(t, os.IsNotExist(fileErr), "File should not exist in dry run mode")
				} else {
					assert.NoError(t, fileErr, "File should exist")

					if fileErr == nil {
						fileContent, readErr := os.ReadFile(expectedFilePath)
						require.NoError(t, readErr)

						var workflow map[string]interface{}
						err := json.Unmarshal(fileContent, &workflow)
						require.NoError(t, err)

						if responseMap, ok := tc.responseBody.(map[string]interface{}); ok {
							if data, ok := responseMap["data"].(n8n.Workflow); ok {
								assert.Equal(t, data.Name, workflow["name"])
								assert.Equal(t, *data.Id, workflow["id"])
								assert.Equal(t, *data.Active, workflow["active"])
							}
						}
					}
				}
			}
		})
	}
}
