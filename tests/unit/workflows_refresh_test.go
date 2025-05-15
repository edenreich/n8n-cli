// Package unit contains unit tests for the n8n-cli
package unit

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshCommand(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name           string
		args           []string
		mockResponses  *n8n.WorkflowList
		mockError      error
		expectedOutput string
		expectError    bool
	}{
		{
			name: "Successfully refreshes workflows",
			args: []string{"--directory", tempDir},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:     stringPtr("123"),
						Name:   "Test Workflow 1",
						Active: boolPtr(true),
					},
					{
						Id:     stringPtr("456"),
						Name:   "Test Workflow 2",
						Active: boolPtr(false),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "Creating workflow 'Test Workflow 1'",
			expectError:    false,
		},
		{
			name:           "Returns error when API call fails",
			args:           []string{"--directory", tempDir},
			mockResponses:  nil,
			mockError:      errors.New("API error"),
			expectedOutput: "error fetching workflows: API error",
			expectError:    true,
		},
		{
			name:           "No workflows found",
			args:           []string{"--directory", tempDir},
			mockResponses:  &n8n.WorkflowList{Data: &[]n8n.Workflow{}},
			mockError:      nil,
			expectedOutput: "No workflows found in n8n instance",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := &clientfakes.FakeClientInterface{}
			fakeClient.GetWorkflowsReturns(tc.mockResponses, tc.mockError)

			viper.Set("api_key", "test_api_key")
			viper.Set("instance_url", "http://test.n8n.local")

			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)

			cmd := &cobra.Command{}
			cmd.Flags().StringP("directory", "d", "", "Directory")
			cmd.Flags().Bool("dry-run", false, "Dry run")
			cmd.Flags().Bool("overwrite", false, "Overwrite")
			cmd.Flags().StringP("output", "o", "json", "Output format")
			cmd.SetOut(outBuf)
			cmd.SetErr(errBuf)

			if err := cmd.Flags().Set("directory", tc.args[1]); err != nil {
				t.Fatal(err)
			}

			directory, _ := cmd.Flags().GetString("directory")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			output, _ := cmd.Flags().GetString("output")

			err := workflows.RefreshWorkflowsWithClient(cmd, fakeClient, directory, dryRun, overwrite, output)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedOutput)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, outBuf.String(), tc.expectedOutput)
			}

			if !tc.expectError && tc.mockResponses != nil && tc.mockResponses.Data != nil && len(*tc.mockResponses.Data) > 0 {
				for _, workflow := range *tc.mockResponses.Data {
					if workflow.Id == nil || *workflow.Id == "" {
						continue
					}

					files, err := os.ReadDir(tempDir)
					require.NoError(t, err)

					found := false
					for _, file := range files {
						if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
							filePath := filepath.Join(tempDir, file.Name())
							content, err := os.ReadFile(filePath)
							require.NoError(t, err)

							if string(content) != "" {
								found = true
								break
							}
						}
					}

					require.True(t, found, "Workflow file should have been created")
				}
			}
		})
	}
}
