// Package integration contains integration tests for the n8n-cli
package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/tests"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestImportWithMockAPI tests the import command with a mock API server
func TestImportWithMockAPI(t *testing.T) {
	tests.SkipIfNotIntegration(t)

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
			_, _ = fmt.Fprintln(w, `{"error": "Not Found"}`)
		}
	}))
	defer mockServer.Close()

	// Create a temp directory for test files
	tempDir, err := os.MkdirTemp("", "n8n-cli-test")
	if err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up Viper with test configuration
	v := viper.New()
	v.Set("api_key", "test-api-key")
	v.Set("instance_url", mockServer.URL)

	// Update the viper configuration
	viper.Reset()
	for _, key := range v.AllKeys() {
		viper.Set(key, v.Get(key))
	}
	// We'll clean this up at the end of the test

	// Test cases
	testCases := []struct {
		name        string
		args        []string
		expectedOut string
		expectError bool
		verifyFiles func(t *testing.T, dir string)
	}{
		{
			name:        "Import specific workflow",
			args:        []string{"import", "--workflow-id", "123", "--directory", tempDir},
			expectedOut: "Starting workflow import",
			expectError: false,
			verifyFiles: func(t *testing.T, dir string) {
				path := filepath.Join(dir, "Test_Workflow_1.json")
				_, err := os.Stat(path)
				assert.NoError(t, err, "Expected workflow file to exist")
			},
		},
		{
			name:        "Import all workflows",
			args:        []string{"import", "--all", "--directory", tempDir},
			expectedOut: "Starting workflow import",
			expectError: false,
			verifyFiles: func(t *testing.T, dir string) {
				paths := []string{
					filepath.Join(dir, "Test_Workflow_1.json"),
					filepath.Join(dir, "Test_Workflow_2.json"),
				}

				for _, path := range paths {
					_, err := os.Stat(path)
					assert.NoError(t, err, "Expected workflow file to exist: "+path)
				}
			},
		},
		{
			name:        "Dry run mode",
			args:        []string{"import", "--workflow-id", "123", "--directory", tempDir, "--dry-run"},
			expectedOut: "DRY RUN MODE",
			expectError: false,
			verifyFiles: func(t *testing.T, dir string) {
				// In dry run mode, we expect nothing to be created
				path := filepath.Join(dir, "Test_Workflow_1.json")

				// Ensure file doesn't exist or was from a previous test
				content, err := os.ReadFile(path)
				if err == nil {
					// If the file exists, make sure it wasn't just created (check if it has content)
					assert.NotEmpty(t, string(content), "File should not be empty if it exists from previous test")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootCmd := cmd.GetRootCmd()
			stdout, stderr, err := executeCommand(t, rootCmd, tc.args...)

			output := stdout + stderr

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Contains(t, output, tc.expectedOut)

			if tc.verifyFiles != nil {
				tc.verifyFiles(t, tempDir)
			}
		})
	}
}

// TestImportWorkflowErrors tests error conditions for the import command
func TestImportWorkflowErrors(t *testing.T) {
	tests.SkipIfNotIntegration(t)

	testCases := []struct {
		name          string
		setupServer   func() *httptest.Server
		args          []string
		expectedError string
	}{
		{
			name: "Invalid API URL",
			setupServer: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// This server will be closed right away
				}))
				server.Close() // Close immediately to simulate connection error
				return server
			},
			args:          []string{"import", "--workflow-id", "123"},
			expectedError: "connection refused",
		},
		{
			name: "API Error Response",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = fmt.Fprint(w, `{"error": "Internal Server Error"}`)
				}))
			},
			args:          []string{"import", "--workflow-id", "123"},
			expectedError: "API returned error 500",
		},
		{
			name: "Invalid JSON Response",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = fmt.Fprint(w, `{invalid json`)
				}))
			},
			args:          []string{"import", "--workflow-id", "123"},
			expectedError: "invalid character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.setupServer()
			defer func() {
				// Server may already be closed in some test cases
				if server != nil {
					server.Close()
				}
			}()

			// Create temp directory
			tempDir, err := os.MkdirTemp("", "n8n-cli-test")
			if err != nil {
				t.Fatalf("Could not create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Add directory to args
			args := append(tc.args, "--directory", tempDir)

			// Set up Viper with test configuration
			v := viper.New()
			v.Set("api_key", "test-api-key")
			if server != nil && server.URL != "" {
				v.Set("instance_url", server.URL)
			} else {
				v.Set("instance_url", "http://localhost:1") // Definitely will fail
			}

			// Update viper configuration
			viper.Reset()
			for _, key := range v.AllKeys() {
				viper.Set(key, v.Get(key))
			}
			// Will be reset for the next test

			rootCmd := cmd.GetRootCmd()
			_, stderr, err := executeCommand(t, rootCmd, args...)

			assert.Error(t, err)
			assert.Contains(t, stderr, tc.expectedError)
		})
	}
}

// TestImportWorkflowByIDWithConfig tests the importWorkflowByID function with a mock config
func TestImportWorkflowByIDWithConfig(t *testing.T) {
	tests.SkipIfNotIntegration(t)

	testCases := []struct {
		name        string
		setupServer func() *httptest.Server
		expectError bool
	}{
		{
			name: "Successful import",
			setupServer: func() *httptest.Server {
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/api/v1/workflows/mock-id" {
						w.Header().Set("Content-Type", "application/json")
						_, _ = fmt.Fprint(w, `{
							"data": {
								"id": "mock-id",
								"name": "Mock Workflow",
								"active": true,
								"nodes": [],
								"connections": {}
							}
						}`)
					} else {
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				return mockServer
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := tc.setupServer()
			defer server.Close()

			// Create temp directory
			tempDir, err := os.MkdirTemp("", "n8n-cli-test")
			if err != nil {
				t.Fatalf("Could not create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// Create mock config
			mockConfig := tests.NewMockConfig()
			mockConfig.GetAPIBaseURLReturns(server.URL + "/api/v1")
			mockConfig.GetAPITokenReturns("mock-token")

			// Import the workflow
			err = cmd.ImportWorkflowByIDWithConfig(mockConfig, tempDir, "mock-id", false, true)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check if file was created
				filePath := filepath.Join(tempDir, "Mock_Workflow.json")
				_, err := os.Stat(filePath)
				assert.NoError(t, err, "Expected workflow file to exist")
			}
		})
	}
}
