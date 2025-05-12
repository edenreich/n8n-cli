package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// trackingTransport is an http.RoundTripper that tracks requests
type trackingTransport struct {
	originalTransport http.RoundTripper
	callback          func(*http.Request, *http.Response)
}

func (t *trackingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.originalTransport.RoundTrip(req)
	if err == nil && t.callback != nil {
		t.callback(req, resp)
	}
	return resp, err
}

// mockServerOptions holds configuration options for the mock server
type mockServerOptions struct {
	workflows   map[string]*n8n.Workflow
	requireAuth bool
	apiToken    string
}

// defaultMockServerOptions provides default configuration for the mock server
func defaultMockServerOptions() *mockServerOptions {
	return &mockServerOptions{
		workflows: map[string]*n8n.Workflow{
			"123": {
				Id:     strPtr("123"),
				Name:   "Test Workflow",
				Active: boolPtr(false),
			},
			"456": {
				Id:     strPtr("456"),
				Name:   "Another Workflow",
				Active: boolPtr(true),
			},
		},
		requireAuth: false,
		apiToken:    "test-token",
	}
}

// setupMockServer creates a test HTTP server that mimics n8n API responses
func setupMockServer(options ...func(*mockServerOptions)) *httptest.Server {
	opts := defaultMockServerOptions()
	for _, option := range options {
		option(opts)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/workflows", func(w http.ResponseWriter, r *http.Request) {
		if opts.requireAuth && !authenticateRequest(w, r, opts.apiToken) {
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetWorkflows(w, r, opts.workflows)
		case http.MethodPost:
			handleCreateWorkflow(w, r, opts.workflows)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `{"error": "Method not allowed"}`)
		}
	})

	mux.HandleFunc("/api/v1/workflows/", func(w http.ResponseWriter, r *http.Request) {
		if opts.requireAuth && !authenticateRequest(w, r, opts.apiToken) {
			return
		}

		pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/"), "/")
		workflowID := pathParts[0]

		if len(pathParts) > 1 {
			action := pathParts[1]
			if r.Method == http.MethodPost && (action == "activate" || action == "deactivate") {
				handleWorkflowActivation(w, r, opts.workflows, workflowID, action == "activate")
				return
			}
		}

		switch r.Method {
		case http.MethodGet:
			handleGetWorkflow(w, r, opts.workflows, workflowID)
		case http.MethodPut:
			handleUpdateWorkflow(w, r, opts.workflows, workflowID)
		case http.MethodDelete:
			handleDeleteWorkflow(w, r, opts.workflows, workflowID)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `{"error": "Method not allowed"}`)
		}
	})

	return httptest.NewServer(mux)
}

// authenticateRequest checks if the request has valid authentication
func authenticateRequest(w http.ResponseWriter, r *http.Request, token string) bool {
	authToken := r.Header.Get("X-N8N-API-KEY")
	if authToken != token {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"error": "Unauthorized"}`)
		return false
	}
	return true
}

// handleGetWorkflows handles GET requests to list workflows
func handleGetWorkflows(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow) {
	query := r.URL.Query().Get("name")
	var matchedWorkflows []*n8n.Workflow

	if query != "" {
		for _, wf := range workflows {
			if strings.Contains(strings.ToLower(wf.Name), strings.ToLower(query)) {
				matchedWorkflows = append(matchedWorkflows, wf)
			}
		}
	} else {
		for _, wf := range workflows {
			matchedWorkflows = append(matchedWorkflows, wf)
		}
	}

	resp := struct {
		Data []*n8n.Workflow `json:"data"`
	}{
		Data: matchedWorkflows,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleGetWorkflow handles GET requests for a single workflow
func handleGetWorkflow(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow, workflowID string) {
	if workflowID == "nonexistent" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "Workflow not found"}`)
		return
	}

	workflow, exists := workflows[workflowID]
	if exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(workflow)
		return
	}

	for _, wf := range workflows {
		if wf.Name == workflowID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(wf)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, `{"error": "Workflow not found"}`)
}

// handleUpdateWorkflow handles PUT requests to update existing workflows
func handleUpdateWorkflow(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow, workflowID string) {
	workflow, exists := workflows[workflowID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "Workflow not found"}`)
		return
	}

	var updatedWorkflow n8n.Workflow
	if err := json.NewDecoder(r.Body).Decode(&updatedWorkflow); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "Invalid workflow data"}`)
		return
	}

	updatedWorkflow.Id = workflow.Id

	workflows[workflowID] = &updatedWorkflow

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedWorkflow)
}

// handleWorkflowActivation handles activation/deactivation requests
func handleWorkflowActivation(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow, workflowID string, activate bool) {
	workflow, exists := workflows[workflowID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "Workflow not found"}`)
		return
	}

	*workflow.Active = activate

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true}`)
}

// handleCreateWorkflow handles POST requests to create new workflows
func handleCreateWorkflow(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow) {
	var workflow n8n.Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "Invalid workflow data"}`)
		return
	}

	if workflow.Id == nil {
		newID := fmt.Sprintf("%d", len(workflows)+1)
		workflow.Id = &newID
	}

	if workflow.Active == nil {
		inactive := false
		workflow.Active = &inactive
	}

	workflows[*workflow.Id] = &workflow

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workflow)
}

// handleDeleteWorkflow handles DELETE requests for a workflow
func handleDeleteWorkflow(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow, workflowID string) {
	_, exists := workflows[workflowID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "Workflow not found"}`)
		return
	}

	delete(workflows, workflowID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true}`)
}

func TestGetServerWorkflows(t *testing.T) {
	server := setupMockServer(func(opts *mockServerOptions) {
		opts.workflows = map[string]*n8n.Workflow{
			"1": {
				Id:     strPtr("1"),
				Name:   "Workflow1",
				Active: boolPtr(true),
			},
			"2": {
				Id:     strPtr("2"),
				Name:   "Workflow2",
				Active: boolPtr(false),
			},
		}
		opts.requireAuth = true
		opts.apiToken = "test-token"
	})
	defer server.Close()

	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer serverError.Close()

	serverInvalidJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer serverInvalidJSON.Close()

	tests := []struct {
		name             string
		apiBaseURL       string
		apiToken         string
		expectedWorkflow []n8n.Workflow
		expectedErr      bool
		errContains      string
	}{
		{
			name:       "Success case - retrieves workflows",
			apiBaseURL: server.URL + "/api/v1",
			apiToken:   "test-token",
			expectedWorkflow: []n8n.Workflow{
				{
					Id:     strPtr("1"),
					Name:   "Workflow1",
					Active: boolPtr(true),
				},
				{
					Id:     strPtr("2"),
					Name:   "Workflow2",
					Active: boolPtr(false),
				},
			},
			expectedErr: false,
		},
		{
			name:        "Error case - HTTP request fails",
			apiBaseURL:  "http://nonexistent-host:12345/api/v1",
			apiToken:    "test-token",
			expectedErr: true,
			errContains: "failed to execute request",
		},
		{
			name:        "Error case - API returns non-200 status",
			apiBaseURL:  serverError.URL,
			apiToken:    "test-token",
			expectedErr: true,
			errContains: "API returned status 500",
		},
		{
			name:        "Error case - Invalid JSON response",
			apiBaseURL:  serverInvalidJSON.URL,
			apiToken:    "test-token",
			expectedErr: true,
			errContains: "failed to parse JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflows, err := getServerWorkflows(tt.apiBaseURL, tt.apiToken)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedWorkflow, workflows)
			}
		})
	}
}

func TestFindWorkflow(t *testing.T) {
	server := setupMockServer(func(opts *mockServerOptions) {
		opts.workflows = map[string]*n8n.Workflow{
			"123": {
				Id:     strPtr("123"),
				Name:   "Test Workflow",
				Active: boolPtr(true),
			},
			"456": {
				Id:     strPtr("456"),
				Name:   "Test Workflow",
				Active: boolPtr(false),
			},
			"111": {
				Id:     strPtr("111"),
				Name:   "Target Workflow",
				Active: boolPtr(true),
			},
			"222": {
				Id:     strPtr("222"),
				Name:   "Common Workflow 1",
				Active: boolPtr(false),
			},
			"333": {
				Id:     strPtr("333"),
				Name:   "Common Workflow 2",
				Active: boolPtr(true),
			},
		}
		opts.requireAuth = true
		opts.apiToken = "test-token"
	})
	defer server.Close()

	tests := []struct {
		name           string
		identifier     string
		apiBaseURL     string
		apiToken       string
		expectedID     string
		expectedName   string
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name:         "Find workflow by ID - success",
			identifier:   "123",
			apiBaseURL:   server.URL + "/api/v1",
			apiToken:     "test-token",
			expectedID:   "123",
			expectedName: "Test Workflow",
		},
		{
			name:         "Find workflow by name - single match",
			identifier:   "Test Workflow",
			apiBaseURL:   server.URL + "/api/v1",
			apiToken:     "test-token",
			expectedID:   "123",
			expectedName: "Test Workflow",
		},
		{
			name:         "Find workflow by name - exact match among multiple",
			identifier:   "Target Workflow",
			apiBaseURL:   server.URL + "/api/v1",
			apiToken:     "test-token",
			expectedID:   "111",
			expectedName: "Target Workflow",
		},
		{
			name:           "No workflow found",
			identifier:     "NonExistent",
			apiBaseURL:     server.URL + "/api/v1",
			apiToken:       "test-token",
			expectedErr:    true,
			expectedErrMsg: "no workflow found with identifier 'NonExistent'",
		},
		{
			name:           "Multiple matches without exact name match",
			identifier:     "Common",
			apiBaseURL:     server.URL + "/api/v1",
			apiToken:       "test-token",
			expectedErr:    true,
			expectedErrMsg: "multiple workflows found matching 'Common'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow, err := findWorkflow(tt.identifier, tt.apiBaseURL, tt.apiToken)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
				assert.Nil(t, workflow)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, workflow)
				assert.Equal(t, tt.expectedID, *workflow.Id)
				assert.Equal(t, tt.expectedName, workflow.Name)
			}
		})
	}
}

func TestSetWorkflowActiveState(t *testing.T) {
	requestsMade := make(map[string]bool)

	server := setupMockServer(func(opts *mockServerOptions) {
		opts.workflows = map[string]*n8n.Workflow{
			"workflow-123": {
				Id:     strPtr("workflow-123"),
				Name:   "Test Workflow",
				Active: boolPtr(false),
			},
		}
		opts.requireAuth = true
		opts.apiToken = "test-token"
	})
	defer server.Close()

	serverWithError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer serverWithError.Close()

	http.DefaultClient.Transport = &trackingTransport{
		originalTransport: http.DefaultTransport,
		callback: func(req *http.Request, resp *http.Response) {
			if strings.Contains(req.URL.Path, "activate") {
				requestsMade["activate"] = true
			} else if strings.Contains(req.URL.Path, "deactivate") {
				requestsMade["deactivate"] = true
			}
		},
	}

	tests := []struct {
		name           string
		workflowID     string
		active         bool
		apiBaseURL     string
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name:        "Activate workflow - success",
			workflowID:  "workflow-123",
			active:      true,
			apiBaseURL:  server.URL + "/api/v1",
			expectedErr: false,
		},
		{
			name:        "Deactivate workflow - success",
			workflowID:  "workflow-123",
			active:      false,
			apiBaseURL:  server.URL + "/api/v1",
			expectedErr: false,
		},
		{
			name:           "Workflow not found",
			workflowID:     "nonexistent",
			active:         true,
			apiBaseURL:     server.URL + "/api/v1",
			expectedErr:    true,
			expectedErrMsg: "no workflow found with identifier 'nonexistent'",
		},
		{
			name:           "Activation API error",
			workflowID:     "workflow-123",
			active:         true,
			apiBaseURL:     serverWithError.URL + "/api/v1",
			expectedErr:    true,
			expectedErrMsg: "API request failed with status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k := range requestsMade {
				delete(requestsMade, k)
			}

			err := setWorkflowActiveState(tt.workflowID, tt.active, tt.apiBaseURL, "test-token")

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)

				endpoint := "activate"
				if !tt.active {
					endpoint = "deactivate"
				}

				assert.True(t, requestsMade[endpoint], "Expected %s request was not made", endpoint)
			}
		})
	}
}

func TestListCommand(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	tests := []struct {
		name           string
		expectedOut    string
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name:        "List workflows successfully",
			expectedOut: "ID   NAME        ACTIVE  \n123  Workflow 1  true    \n456  Workflow 2  false   \n",
			expectedErr: false,
		},
		{
			name:           "Error fetching workflows",
			expectedErr:    true,
			expectedErrMsg: "failed to get workflows",
		},
		{
			name:        "Empty workflow list",
			expectedOut: "ID  NAME  ACTIVE  \n",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			listCmd.SetOut(buf)
			listCmd.SetErr(buf)
			err := listCmd.Execute()

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOut, buf.String())
			}
		})
	}
}

func TestActivateCommand(t *testing.T) {
	server := setupMockServer(func(opts *mockServerOptions) {
		opts.requireAuth = true
	})
	defer server.Close()

	// Save original environment
	originalEnvVars := map[string]string{
		"N8N_INSTANCE_URL": os.Getenv("N8N_INSTANCE_URL"),
		"N8N_API_KEY":      os.Getenv("N8N_API_KEY"),
	}

	// Restore original environment after test
	defer func() {
		for k, v := range originalEnvVars {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	// Set environment variables for the test
	os.Setenv("N8N_INSTANCE_URL", server.URL)
	os.Setenv("N8N_API_KEY", "test-token")

	// Reset viper to pick up the new environment variables
	viper.Reset()
	initConfig()

	tests := []struct {
		name           string
		args           []string
		expectedOut    string
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name:        "Activate workflow successfully",
			args:        []string{"123"},
			expectedOut: "Workflow '123' has been activated\n",
			expectedErr: false,
		},
		{
			name:           "Activate non-existent workflow",
			args:           []string{"nonexistent"},
			expectedErr:    true,
			expectedErrMsg: "failed to activate workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			activateCmd.SetOut(buf)
			activateCmd.SetErr(buf)
			activateCmd.SetArgs(tt.args)
			err := activateCmd.Execute()

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOut, buf.String())
			}
		})
	}
}

func TestDeactivateCommand(t *testing.T) {
	server := setupMockServer(func(opts *mockServerOptions) {
		opts.requireAuth = true
	})
	defer server.Close()

	tests := []struct {
		name           string
		args           []string
		expectedOut    string
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name:        "Deactivate workflow successfully",
			args:        []string{"123"},
			expectedOut: "Workflow '123' has been deactivated\n",
			expectedErr: false,
		},
		{
			name:           "Deactivate non-existent workflow",
			args:           []string{"nonexistent"},
			expectedErr:    true,
			expectedErrMsg: "failed to deactivate workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			deactivateCmd.SetOut(buf)
			deactivateCmd.SetErr(buf)
			deactivateCmd.SetArgs(tt.args)
			err := deactivateCmd.Execute()

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOut, buf.String())
			}
		})
	}
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
