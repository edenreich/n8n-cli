// Package unit contains unit tests for the n8n-cli
package unit

import (
	"bytes"
	"encoding/json"
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
	"gopkg.in/yaml.v3"
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
		setupFiles     func(t *testing.T, dir string)
	}{
		{
			name: "Successfully refreshes workflows (JSON format)",
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
			name: "Successfully refreshes workflows (YAML format)",
			args: []string{"--directory", tempDir, "--output", "yaml"},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:     stringPtr("789"),
						Name:   "Test Workflow 3",
						Active: boolPtr(true),
					},
					{
						Id:     stringPtr("abc"),
						Name:   "Test Workflow 4",
						Active: boolPtr(false),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "Creating workflow 'Test Workflow 3'",
			expectError:    false,
		},
		{
			name: "Detects no changes when content is identical (JSON)",
			args: []string{"--directory", tempDir},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:     stringPtr("123"),
						Name:   "Test Workflow 1",
						Active: boolPtr(true),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "No changes for workflow",
			expectError:    false,
			setupFiles: func(t *testing.T, dir string) {
				workflow := n8n.Workflow{
					Id:     stringPtr("123"),
					Name:   "Test Workflow 1",
					Active: boolPtr(true),
				}
				content, err := json.MarshalIndent(workflow, "", "  ")
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "Test_Workflow_1.json"), content, 0644)
				require.NoError(t, err)
			},
		},
		{
			name: "Detects no changes when content is identical (YAML)",
			args: []string{"--directory", tempDir, "--output", "yaml"},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:     stringPtr("789"),
						Name:   "Test Workflow 3",
						Active: boolPtr(true),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "No changes for workflow",
			expectError:    false,
			setupFiles: func(t *testing.T, dir string) {
				workflow := n8n.Workflow{
					Id:     stringPtr("789"),
					Name:   "Test Workflow 3",
					Active: boolPtr(true),
				}
				content, err := yaml.Marshal(workflow)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "Test_Workflow_3.yaml"), content, 0644)
				require.NoError(t, err)
			},
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
			testDir := filepath.Join(tempDir, t.Name())
			err := os.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			if tc.setupFiles != nil {
				tc.setupFiles(t, testDir)
			}

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

			if err := cmd.Flags().Set("directory", testDir); err != nil {
				t.Fatal(err)
			}

			for i := 0; i < len(tc.args); i++ {
				if tc.args[i] == "--output" || tc.args[i] == "-o" {
					if i+1 < len(tc.args) {
						if err := cmd.Flags().Set("output", tc.args[i+1]); err != nil {
							t.Fatal(err)
						}
						break
					}
				}
			}

			directory, _ := cmd.Flags().GetString("directory")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			output, _ := cmd.Flags().GetString("output")

			err = workflows.RefreshWorkflowsWithClient(cmd, fakeClient, directory, dryRun, overwrite, output)

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

					files, err := os.ReadDir(directory)
					require.NoError(t, err)

					found := false
					expectedExt := ".json"
					if cmd.Flags().Changed("output") {
						outputFormat, _ := cmd.Flags().GetString("output")
						if outputFormat == "yaml" || outputFormat == "yml" {
							expectedExt = ".yaml"
						}
					}

					for _, file := range files {
						if !file.IsDir() && (filepath.Ext(file.Name()) == expectedExt) {
							filePath := filepath.Join(directory, file.Name())
							content, err := os.ReadFile(filePath)
							require.NoError(t, err)

							if string(content) != "" {
								found = true
								break
							}
						}
					}

					require.True(t, found, "Workflow file should have been created with %s extension", expectedExt)
				}
			}
		})
	}
}
