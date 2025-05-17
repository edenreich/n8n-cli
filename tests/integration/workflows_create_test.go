// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkflow(t *testing.T) {
	testCases := []struct {
		name           string
		workflow       n8n.Workflow
		expectedError  bool
		errorContains  string
		validateFields func(t *testing.T, sentWorkflow map[string]interface{})
	}{
		{
			name: "Create workflow without ID and active fields",
			workflow: n8n.Workflow{
				Name: "Test Workflow",
			},
			expectedError: false,
			validateFields: func(t *testing.T, sentWorkflow map[string]interface{}) {
				_, hasID := sentWorkflow["id"]
				_, hasActive := sentWorkflow["active"]
				assert.False(t, hasID, "ID field should be excluded when creating workflow")
				assert.False(t, hasActive, "active field should be excluded when creating workflow")
			},
		},
		{
			name: "Create workflow with ID and active fields",
			workflow: func() n8n.Workflow {
				id := "test-id"
				active := true
				return n8n.Workflow{
					Id:     &id,
					Name:   "Test Workflow With Fields",
					Active: &active,
				}
			}(),
			expectedError: false,
			validateFields: func(t *testing.T, sentWorkflow map[string]interface{}) {
				_, hasID := sentWorkflow["id"]
				_, hasActive := sentWorkflow["active"]
				assert.False(t, hasID, "ID field should be excluded when creating workflow")
				assert.False(t, hasActive, "active field should be excluded when creating workflow")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var receivedBody map[string]interface{}

			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodPost {
					err := json.NewDecoder(r.Body).Decode(&receivedBody)
					require.NoError(t, err, "Failed to decode request body")

					w.Header().Set("Content-Type", "application/json")
					newID := "new-id"
					resp := n8n.Workflow{
						Id:   &newID,
						Name: tc.workflow.Name,
					}
					err = json.NewEncoder(w).Encode(resp)
					require.NoError(t, err, "Failed to encode response")
					return
				}

				w.WriteHeader(http.StatusNotFound)
			}))
			defer mockServer.Close()

			viper.Reset()
			viper.Set("api_key", "test-api-key")
			viper.Set("instance_url", mockServer.URL)

			client := n8n.NewClient(mockServer.URL, "test-api-key")
			result, err := client.CreateWorkflow(&tc.workflow)

			if tc.expectedError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "new-id", *result.Id)
			}

			tc.validateFields(t, receivedBody)
		})
	}
}

func TestWorkflowSyncBug14(t *testing.T) {
	id := "test-id"
	active := true
	testWorkflow := n8n.Workflow{
		Id:     &id,
		Name:   "Test Workflow With Fields",
		Active: &active,
	}

	var requestReceived bool
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodPost {
			var receivedWorkflow map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&receivedWorkflow)
			require.NoError(t, err, "Failed to decode request body")

			_, hasID := receivedWorkflow["id"]
			_, hasActive := receivedWorkflow["active"]

			if hasID || hasActive {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprintln(w, `{"error":"body/active is read only"}`)
				return
			}

			requestReceived = true
			w.Header().Set("Content-Type", "application/json")
			newID := "server-generated-id"
			resp := n8n.Workflow{
				Id:   &newID,
				Name: testWorkflow.Name,
			}
			err = json.NewEncoder(w).Encode(resp)
			require.NoError(t, err, "Failed to encode response")
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	client := n8n.NewClient(mockServer.URL, "test-api-key")
	result, err := client.CreateWorkflow(&testWorkflow)

	assert.NoError(t, err, "Creating workflow should not error with ID and active fields")
	assert.NotNil(t, result)
	assert.Equal(t, "server-generated-id", *result.Id)
	assert.True(t, requestReceived, "Request to server was not received")
}
