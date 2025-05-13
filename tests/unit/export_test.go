// Package unit contains unit tests for the n8n-cli
package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/stretchr/testify/assert"
)

func TestSetWorkflowActiveState(t *testing.T) {
	testCases := []struct {
		name           string
		workflowID     string
		activate       bool
		responseStatus int
		expectedErr    bool
	}{
		{
			name:           "Activate workflow - Success",
			workflowID:     "123",
			activate:       true,
			responseStatus: http.StatusOK,
			expectedErr:    false,
		},
		{
			name:           "Deactivate workflow - Success",
			workflowID:     "456",
			activate:       false,
			responseStatus: http.StatusOK,
			expectedErr:    false,
		},
		{
			name:           "Error - Server error",
			workflowID:     "789",
			activate:       true,
			responseStatus: http.StatusInternalServerError,
			expectedErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)

				expectedAction := "activate"
				if !tc.activate {
					expectedAction = "deactivate"
				}

				expectedPath := "/api/v1/workflows/" + tc.workflowID + "/" + expectedAction
				assert.Equal(t, expectedPath, r.URL.Path)

				assert.Equal(t, "test-api-key", r.Header.Get("X-N8N-API-KEY"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(tc.responseStatus)
			}))
			defer mockServer.Close()

			baseURL := cmd.FormatAPIBaseURL(mockServer.URL)
			err := cmd.SetWorkflowActiveState(baseURL, "test-api-key", tc.workflowID, tc.activate)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetServerWorkflows(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		responseStatus int
		responseBody   interface{}
		expectedErr    bool
		expectedCount  int
	}{
		{
			name:           "Success - Multiple workflows",
			responseStatus: http.StatusOK,
			responseBody: n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:     stringPtr("1"),
						Name:   "First Workflow",
						Active: boolPtr(true),
					},
					{
						Id:     stringPtr("2"),
						Name:   "Second Workflow",
						Active: boolPtr(false),
					},
				},
			},
			expectedErr:   false,
			expectedCount: 2,
		},
		{
			name:           "Success - Empty workflows",
			responseStatus: http.StatusOK,
			responseBody: n8n.WorkflowList{
				Data: &[]n8n.Workflow{},
			},
			expectedErr:   false,
			expectedCount: 0,
		},
		{
			name:           "Error - API error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   nil,
			expectedErr:    true,
			expectedCount:  0,
		},
		{
			name:           "Error - Invalid response format",
			responseStatus: http.StatusOK,
			responseBody:   "invalid json",
			expectedErr:    true,
			expectedCount:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/workflows", r.URL.Path)

				assert.Equal(t, "test-api-key", r.Header.Get("X-N8N-API-KEY"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(tc.responseStatus)

				if tc.responseBody != nil {
					if s, ok := tc.responseBody.(string); ok {
						_, _ = w.Write([]byte(s))
					} else {
						respBytes, _ := json.Marshal(tc.responseBody)
						_, _ = w.Write(respBytes)
					}
				}
			}))
			defer mockServer.Close()

			baseURL := cmd.FormatAPIBaseURL(mockServer.URL)
			workflows, err := cmd.GetServerWorkflows(baseURL, "test-api-key")

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, workflows, tc.expectedCount)

				if tc.expectedCount > 0 {
					workflowList := tc.responseBody.(n8n.WorkflowList)
					expectedWorkflows := *workflowList.Data

					for i, wf := range workflows {
						expected := expectedWorkflows[i]
						assert.Equal(t, *expected.Id, *wf.Id)
						assert.Equal(t, expected.Name, wf.Name)
						assert.Equal(t, *expected.Active, *wf.Active)
					}
				}
			}
		})
	}
}
