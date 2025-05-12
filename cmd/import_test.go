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
	"testing"

	"github.com/stretchr/testify/assert"
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

			if tc.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Did not expect error but got one")
			}

			output := stdout + stderr
			assert.Contains(t, output, tc.expectedOut, "Expected output to contain specific text")
		})
	}
}

func TestImportWithMockAPI(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-N8N-API-KEY")
		if apiKey != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		switch r.URL.Path {
		case "/api/v1/workflows":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{
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
			_, _ = fmt.Fprint(w, `{
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
			_, _ = fmt.Fprint(w, `{
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
			_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
		}
	}))
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "n8n-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp directory: %v", err)
		}
	}()

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
		assert.NoError(t, err, "Failed to import workflow")

		filePath := filepath.Join(tempDir, "Test_Workflow_1.json")
		_, err = os.Stat(filePath)
		assert.False(t, os.IsNotExist(err), "Expected workflow file %s was not created", filePath)

		content, err := os.ReadFile(filePath)
		assert.NoError(t, err, "Failed to read workflow file")

		var workflow map[string]interface{}
		err = json.Unmarshal(content, &workflow)
		assert.NoError(t, err, "Workflow file does not contain valid JSON")

		id, ok := workflow["id"].(string)
		assert.True(t, ok && id == "123", "Workflow has incorrect ID: %v", workflow["id"])

		name, ok := workflow["name"].(string)
		assert.True(t, ok && name == "Test Workflow 1", "Workflow has incorrect name: %v", workflow["name"])
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

		_ = os.RemoveAll(tempDir)
		_ = os.MkdirAll(tempDir, 0755)

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
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Logf("Failed to remove temp directory: %v", err)
			}
		}()

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
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Logf("Failed to clean up temp directory: %v", err)
			}
		}()

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, `{"error": "Server error"}`)
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
		assert.Error(t, err, "Expected error with server error but got none")
	})

	t.Run("Invalid JSON Response", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "n8n-cli-test-error")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Logf("Failed to clean up temp directory: %v", err)
			}
		}()

		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{invalid json}`)
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
		assert.Error(t, err, "Expected error with invalid JSON but got none")
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

func TestImportWorkflowByIDWithConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "n8n-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to clean up temp directory: %v", err)
		}
	}()

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
			_, _ = fmt.Fprint(w, responseBody)
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
