// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/edenreich/n8n-cli/n8n"
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
			_, err := fmt.Fprintf(w, `{"error": "Method not allowed"}`)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
			}

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
		case http.MethodPut, http.MethodPatch:
			handleUpdateWorkflow(w, r, opts.workflows, workflowID)
		case http.MethodDelete:
			handleDeleteWorkflow(w, r, opts.workflows, workflowID)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, err := fmt.Fprintf(w, `{"error": "Method not allowed"}`)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
			}
		}
	})

	return httptest.NewServer(mux)
}

// authenticateRequest validates the API token
func authenticateRequest(w http.ResponseWriter, r *http.Request, expectedToken string) bool {
	token := r.Header.Get("X-N8N-API-KEY")
	if token != expectedToken {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := fmt.Fprintf(w, `{"error": "Unauthorized"}`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
		}
		return false
	}
	return true
}

// handleGetWorkflows returns all workflows
func handleGetWorkflows(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	data := make([]*n8n.Workflow, 0, len(workflows))
	for _, wf := range workflows {
		data = append(data, wf)
	}

	response := map[string]interface{}{
		"data": data,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
	}
}

// handleGetWorkflow returns a specific workflow
func handleGetWorkflow(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow, workflowID string) {
	workflow, exists := workflows[workflowID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := fmt.Fprintf(w, `{"error": "Workflow not found"}`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"data": workflow,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
	}
}

// handleCreateWorkflow creates a new workflow
func handleCreateWorkflow(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow) {
	var workflow n8n.Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := fmt.Fprintf(w, `{"error": "Invalid request body"}`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
		}
		return
	}

	// Create a new ID if not provided
	if workflow.Id == nil {
		id := "new-id"
		workflow.Id = &id
	}

	workflows[*workflow.Id] = &workflow

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"data": workflow,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
	}
}

// handleUpdateWorkflow updates an existing workflow
func handleUpdateWorkflow(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow, workflowID string) {
	_, exists := workflows[workflowID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := fmt.Fprintf(w, `{"error": "Workflow not found"}`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
		}
		return
	}

	var updatedWorkflow n8n.Workflow
	if err := json.NewDecoder(r.Body).Decode(&updatedWorkflow); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := fmt.Fprintf(w, `{"error": "Invalid request body"}`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
		}
		return
	}

	updatedWorkflow.Id = &workflowID
	workflows[workflowID] = &updatedWorkflow

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"data": updatedWorkflow,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
	}
}

// handleDeleteWorkflow deletes a workflow
func handleDeleteWorkflow(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow, workflowID string) {
	_, exists := workflows[workflowID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := fmt.Fprintf(w, `{"error": "Workflow not found"}`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
		}
		return
	}

	delete(workflows, workflowID)

	w.WriteHeader(http.StatusNoContent)
}

// handleWorkflowActivation activates or deactivates a workflow
func handleWorkflowActivation(w http.ResponseWriter, r *http.Request, workflows map[string]*n8n.Workflow, workflowID string, activate bool) {
	workflow, exists := workflows[workflowID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, err := fmt.Fprintf(w, `{"error": "Workflow not found"}`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
		}
		return
	}

	workflow.Active = &activate

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"data": workflow,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
	}
}
