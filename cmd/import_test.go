package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportCommand(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
		setupEnv    func() string
		cleanup     func(string)
	}{
		{
			name:        "Help output",
			args:        []string{"import", "--help"},
			expectedOut: "Import command fetches workflows from your n8n instance",
			expectError: false,
			setupEnv:    func() string { return "" },
			cleanup:     func(string) {},
		},
		{
			name:        "Dry run mode",
			args:        []string{"import", "--dry-run"},
			expectedOut: "Show what would be done without making changes",
			expectError: false,
			setupEnv:    func() string { return "" },
			cleanup:     func(string) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := tc.setupEnv()
			defer tc.cleanup(tempDir)

			cmd := GetRootCmd()

			stdout, stderr, err := executeCommand(cmd, tc.args...)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			output := stdout + stderr
			if !strings.Contains(output, tc.expectedOut) {
				t.Errorf("Expected output to contain '%s', got:\n%s", tc.expectedOut, output)
			}
		})
	}
}

func TestImportWithMockAPI(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-N8N-API-KEY")
		if apiKey != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		switch r.URL.Path {
		case "/api/v1/workflows":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
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
					}
				]
			}`)

		case "/api/v1/workflows/123":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"data": {
					"id": "123",
					"name": "Test Workflow 1",
					"active": true,
					"nodes": [],
					"connections": {}
				}
			}`)

		case "/api/v1/workflows/456":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"data": {
					"id": "456",
					"name": "Test Workflow 2",
					"active": false,
					"nodes": [],
					"connections": {}
				}
			}`)

		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error": "Not found"}`)
		}
	}))
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "n8n-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Import specific workflow", func(t *testing.T) {
		config := ImportConfig{
			Directory:  tempDir,
			APIBaseURL: mockServer.URL + "/api/v1",
			APIToken:   "test-api-key",
			WorkflowID: "123",
			All:        false,
			DryRun:     false,
			Verbose:    true,
		}

		err := importWorkflows(config)
		if err != nil {
			t.Errorf("Failed to import workflow: %v", err)
		}

		filePath := filepath.Join(tempDir, "Test_Workflow_1.json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected workflow file %s was not created", filePath)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read workflow file: %v", err)
		}

		var workflow map[string]interface{}
		if err := json.Unmarshal(content, &workflow); err != nil {
			t.Errorf("Workflow file does not contain valid JSON: %v", err)
		}

		id, ok := workflow["id"].(string)
		if !ok || id != "123" {
			t.Errorf("Workflow has incorrect ID: %v", workflow["id"])
		}

		name, ok := workflow["name"].(string)
		if !ok || name != "Test Workflow 1" {
			t.Errorf("Workflow has incorrect name: %v", workflow["name"])
		}
	})

	t.Run("Import all workflows", func(t *testing.T) {
		config := ImportConfig{
			Directory:  tempDir,
			APIBaseURL: mockServer.URL + "/api/v1",
			APIToken:   "test-api-key",
			WorkflowID: "",
			All:        true,
			DryRun:     false,
			Verbose:    false,
		}

		err := importWorkflows(config)
		if err != nil {
			t.Errorf("Failed to import workflows: %v", err)
		}

		filePaths := []string{
			filepath.Join(tempDir, "Test_Workflow_1.json"),
			filepath.Join(tempDir, "Test_Workflow_2.json"),
		}

		for _, filePath := range filePaths {
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Expected workflow file %s was not created", filePath)
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Errorf("Failed to read workflow file %s: %v", filePath, err)
				continue
			}

			var workflow map[string]interface{}
			if err := json.Unmarshal(content, &workflow); err != nil {
				t.Errorf("Workflow file %s does not contain valid JSON: %v", filePath, err)
			}
		}
	})

	t.Run("Dry run mode", func(t *testing.T) {
		config := ImportConfig{
			Directory:  tempDir,
			APIBaseURL: mockServer.URL + "/api/v1",
			APIToken:   "test-api-key",
			WorkflowID: "123",
			All:        false,
			DryRun:     true,
			Verbose:    true,
		}

		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, 0755)

		err := importWorkflows(config)
		if err != nil {
			t.Errorf("Failed to import workflow in dry run mode: %v", err)
		}

		files, err := os.ReadDir(tempDir)
		if err != nil {
			t.Errorf("Failed to read temp directory: %v", err)
		}
		if len(files) > 0 {
			t.Errorf("Expected no files in directory, but found %d files", len(files))
		}
	})
}

func TestImportWorkflowErrors(t *testing.T) {
	t.Run("Invalid API URL", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "n8n-cli-test-error")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		config := ImportConfig{
			Directory:  tempDir,
			APIBaseURL: "http://invalid-url-that-does-not-exist.local/api/v1",
			APIToken:   "test-api-key",
			WorkflowID: "123",
			All:        false,
			DryRun:     false,
			Verbose:    false,
		}

		err = importWorkflows(config)
		if err == nil {
			t.Errorf("Expected error with invalid API URL but got none")
		}
	})

	t.Run("API Error Response", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "n8n-cli-test-error")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error": "Server error"}`)
		}))
		defer mockServer.Close()

		config := ImportConfig{
			Directory:  tempDir,
			APIBaseURL: mockServer.URL + "/api/v1",
			APIToken:   "test-api-key",
			WorkflowID: "123",
			All:        false,
			DryRun:     false,
			Verbose:    false,
		}

		err = importWorkflows(config)
		if err == nil {
			t.Errorf("Expected error with server error but got none")
		}
	})

	t.Run("Invalid JSON Response", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "n8n-cli-test-error")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{invalid json}`)
		}))
		defer mockServer.Close()

		config := ImportConfig{
			Directory:  tempDir,
			APIBaseURL: mockServer.URL + "/api/v1",
			APIToken:   "test-api-key",
			WorkflowID: "",
			All:        true,
			DryRun:     false,
			Verbose:    false,
		}

		err = importWorkflows(config)
		if err == nil {
			t.Errorf("Expected error with invalid JSON but got none")
		}
	})
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple Name", "Simple_Name"},
		{"Name/With/Slashes", "Name_With_Slashes"},
		{"Name\\With\\Backslashes", "Name_With_Backslashes"},
		{"Name:With:Colons", "Name_With_Colons"},
		{"Name*With*Stars", "Name_With_Stars"},
		{"Name?With?Questions", "Name_With_Questions"},
		{"Name\"With\"Quotes", "Name_With_Quotes"},
		{"Name<With>Brackets", "Name_With_Brackets"},
		{"Name|With|Pipes", "Name_With_Pipes"},
		{"Complex: Name*/\\?\"<>|", "Complex__Name________"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := sanitizeFilename(test.input)
			if result != test.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", test.input, result, test.expected)
			}
		})
	}
}

// Mock HTTP client for testing API interactions
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestImportWorkflowByIDWithConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "n8n-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	responseBody := `{
		"data": {
			"id": "mock-id",
			"name": "Mock Workflow",
			"active": true,
			"nodes": []
		}
	}`

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
		Header:     make(http.Header),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	oldClient := http.DefaultClient
	http.DefaultClient = httpClient

	t.Run("Successful import", func(t *testing.T) {
		config := ImportConfig{
			Directory:  tempDir,
			APIBaseURL: "http://localhost/api/v1",
			APIToken:   "mock-token",
			WorkflowID: "mock-id",
			All:        false,
			DryRun:     false,
			Verbose:    true,
		}

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, responseBody)
		}))
		defer mockServer.Close()

		config.APIBaseURL = mockServer.URL + "/api/v1"

		err := importWorkflowByIDWithConfig("mock-id", config)
		if err != nil {
			t.Errorf("Failed to import workflow: %v", err)
		}

		filePath := filepath.Join(tempDir, "Mock_Workflow.json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected workflow file %s was not created", filePath)
		}
	})

	http.DefaultClient = oldClient
}

// Helper type for mocking HTTP transport
type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
