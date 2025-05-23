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
		validateFiles  func(t *testing.T, dir string)
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
			name: "Verify YAML files have separator",
			args: []string{"--directory", tempDir, "--output", "yaml"},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:     stringPtr("yaml-separator"),
						Name:   "YAML Separator Test",
						Active: boolPtr(true),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "Creating workflow 'YAML Separator Test'",
			expectError:    false,
			validateFiles: func(t *testing.T, dir string) {
				filePath := filepath.Join(dir, "YAML_Separator_Test.yaml")
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)

				assert.True(t, bytes.HasPrefix(content, []byte("---\n")),
					"YAML file should start with '---' separator")
			},
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
				filePath := filepath.Join(dir, "Test_Workflow_3.yaml")

				yamlContent := []byte(`---
id: "789"
name: "Test Workflow 3"
active: true
`)
				err := os.WriteFile(filePath, yamlContent, 0644)
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
		{
			name: "Excludes createdAt and updatedAt fields",
			args: []string{"--directory", tempDir},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:        stringPtr("abc123"),
						Name:      "Workflow With Timestamps",
						Active:    boolPtr(true),
						CreatedAt: timePtr("2025-05-11T18:58:01.685Z"),
						UpdatedAt: timePtr("2025-05-14T23:48:45.83Z"),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "Creating workflow 'Workflow With Timestamps'",
			expectError:    false,
			validateFiles: func(t *testing.T, dir string) {
				filePath := filepath.Join(dir, "Workflow_With_Timestamps.json")
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)

				var workflow map[string]interface{}
				err = json.Unmarshal(content, &workflow)
				require.NoError(t, err)

				_, hasCreatedAt := workflow["createdAt"]
				_, hasUpdatedAt := workflow["updatedAt"]
				assert.False(t, hasCreatedAt, "createdAt should be excluded from the workflow")
				assert.False(t, hasUpdatedAt, "updatedAt should be excluded from the workflow")
			},
		},
		{
			name: "Excludes createdAt and updatedAt fields",
			args: []string{"--directory", tempDir},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:        stringPtr("abc123"),
						Name:      "Workflow With Timestamps",
						Active:    boolPtr(true),
						CreatedAt: timePtr("2025-05-11T18:58:01.685Z"),
						UpdatedAt: timePtr("2025-05-14T23:48:45.83Z"),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "Creating workflow 'Workflow With Timestamps'",
			expectError:    false,
			validateFiles: func(t *testing.T, dir string) {
				filePath := filepath.Join(dir, "Workflow_With_Timestamps.json")
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)

				var workflow map[string]interface{}
				err = json.Unmarshal(content, &workflow)
				require.NoError(t, err)

				_, hasCreatedAt := workflow["createdAt"]
				_, hasUpdatedAt := workflow["updatedAt"]
				assert.False(t, hasCreatedAt, "createdAt should be excluded from the workflow")
				assert.False(t, hasUpdatedAt, "updatedAt should be excluded from the workflow")
			},
		},
		{
			name: "Only refreshes existing workflows when --all is not specified",
			args: []string{"--directory", tempDir},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:     stringPtr("existing123"),
						Name:   "Existing Workflow",
						Active: boolPtr(true),
					},
					{
						Id:     stringPtr("nonexisting456"),
						Name:   "Non-Existing Workflow",
						Active: boolPtr(false),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "Refreshing only workflows that exist in the directory",
			expectError:    false,
			setupFiles: func(t *testing.T, dir string) {
				// Create a file for the "existing123" workflow
				filePath := filepath.Join(dir, "Existing_Workflow.json")
				workflow := n8n.Workflow{
					Id:     stringPtr("existing123"),
					Name:   "Existing Workflow",
					Active: boolPtr(false), // We'll set it to false so we can detect the update
				}

				encoder := n8n.NewWorkflowEncoder(true)
				content, err := encoder.EncodeToJSON(workflow)
				require.NoError(t, err)

				err = os.WriteFile(filePath, content, 0644)
				require.NoError(t, err)
			},
			validateFiles: func(t *testing.T, dir string) {
				// The existing workflow should be updated
				existingPath := filepath.Join(dir, "Existing_Workflow.json")
				content, err := os.ReadFile(existingPath)
				require.NoError(t, err)

				var workflow map[string]interface{}
				err = json.Unmarshal(content, &workflow)
				require.NoError(t, err)

				assert.Equal(t, true, workflow["active"], "Existing workflow should be updated to active=true")

				// The non-existing workflow should not be created
				nonExistingPath := filepath.Join(dir, "Non-Existing_Workflow.json")
				_, err = os.Stat(nonExistingPath)
				assert.True(t, os.IsNotExist(err), "Non-existing workflow should not be created without --all flag")
			},
		},
		{
			name: "Refreshes all workflows when --all flag is specified",
			args: []string{"--directory", tempDir, "--all"},
			mockResponses: &n8n.WorkflowList{
				Data: &[]n8n.Workflow{
					{
						Id:     stringPtr("existing123"),
						Name:   "Existing Workflow",
						Active: boolPtr(true),
					},
					{
						Id:     stringPtr("nonexisting456"),
						Name:   "Non-Existing Workflow",
						Active: boolPtr(true),
					},
				},
			},
			mockError:      nil,
			expectedOutput: "Refreshing all workflows from n8n instance",
			expectError:    false,
			setupFiles: func(t *testing.T, dir string) {
				filePath := filepath.Join(dir, "Existing_Workflow.json")
				workflow := n8n.Workflow{
					Id:     stringPtr("existing123"),
					Name:   "Existing Workflow",
					Active: boolPtr(false),
				}

				encoder := n8n.NewWorkflowEncoder(true)
				content, err := encoder.EncodeToJSON(workflow)
				require.NoError(t, err)

				err = os.WriteFile(filePath, content, 0644)
				require.NoError(t, err)
			},
			validateFiles: func(t *testing.T, dir string) {
				existingPath := filepath.Join(dir, "Existing_Workflow.json")
				content, err := os.ReadFile(existingPath)
				require.NoError(t, err)

				var workflow map[string]interface{}
				err = json.Unmarshal(content, &workflow)
				require.NoError(t, err)

				assert.Equal(t, true, workflow["active"], "Existing workflow should be updated to active=true")

				nonExistingPath := filepath.Join(dir, "Non-Existing_Workflow.json")
				content, err = os.ReadFile(nonExistingPath)
				require.NoError(t, err)

				err = json.Unmarshal(content, &workflow)
				require.NoError(t, err)

				assert.Equal(t, "nonexisting456", workflow["id"], "Non-existing workflow should be created with --all flag")
				assert.Equal(t, true, workflow["active"], "Non-existing workflow should have active=true")
			},
		},
		{
			name:           "Handles errors when fetching individual workflows",
			args:           []string{"--directory", tempDir},
			mockResponses:  nil,
			mockError:      nil,
			expectedOutput: "Warning: Could not fetch workflow with ID",
			expectError:    false,
			setupFiles: func(t *testing.T, dir string) {
				filePath := filepath.Join(dir, "Error_Workflow.json")
				workflow := n8n.Workflow{
					Id:     stringPtr("error123"),
					Name:   "Error Workflow",
					Active: boolPtr(false),
				}

				encoder := n8n.NewWorkflowEncoder(true)
				content, err := encoder.EncodeToJSON(workflow)
				require.NoError(t, err)

				err = os.WriteFile(filePath, content, 0644)
				require.NoError(t, err)

				successPath := filepath.Join(dir, "Success_Workflow.json")
				successWorkflow := n8n.Workflow{
					Id:     stringPtr("success456"),
					Name:   "Success Workflow",
					Active: boolPtr(false),
				}

				content, err = encoder.EncodeToJSON(successWorkflow)
				require.NoError(t, err)

				err = os.WriteFile(successPath, content, 0644)
				require.NoError(t, err)
			},
			validateFiles: func(t *testing.T, dir string) {
				successPath := filepath.Join(dir, "Success_Workflow.json")
				content, err := os.ReadFile(successPath)
				require.NoError(t, err)

				var workflow map[string]interface{}
				err = json.Unmarshal(content, &workflow)
				require.NoError(t, err)

				assert.Equal(t, true, workflow["active"], "Success workflow should be updated to active=true")
			},
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

			fakeClient.GetWorkflowCalls(func(id string) (*n8n.Workflow, error) {
				if tc.name == "Handles errors when fetching individual workflows" {
					switch id {
					case "success456":
						return &n8n.Workflow{
							Id:     stringPtr("success456"),
							Name:   "Success Workflow",
							Active: boolPtr(true),
						}, nil
					case "error123":
						return nil, errors.New("API error fetching workflow")
					}
				}

				if tc.mockResponses != nil && tc.mockResponses.Data != nil {
					for _, workflow := range *tc.mockResponses.Data {
						if workflow.Id != nil && *workflow.Id == id {
							return &workflow, nil
						}
					}
				}

				return nil, errors.New("workflow not found")
			})

			viper.Set("api_key", "test_api_key")
			viper.Set("instance_url", "http://test.n8n.local")

			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)

			cmd := &cobra.Command{}
			cmd.Flags().StringP("directory", "d", "", "Directory")
			cmd.Flags().Bool("dry-run", false, "Dry run")
			cmd.Flags().Bool("overwrite", false, "Overwrite")
			cmd.Flags().StringP("output", "o", "json", "Output format")
			cmd.Flags().Bool("no-truncate", false, "Include all fields in output")
			cmd.Flags().Bool("all", false, "Refresh all workflows")
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
			noTruncate, _ := cmd.Flags().GetBool("no-truncate")
			minimal := !noTruncate

			all := false
			for i := 0; i < len(tc.args); i++ {
				if tc.args[i] == "--all" {
					all = true
					break
				}
			}

			err = workflows.RefreshWorkflowsWithClient(cmd, fakeClient, directory, dryRun, overwrite, output, minimal, all)

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

			if tc.validateFiles != nil {
				tc.validateFiles(t, directory)
			}
		})
	}
}
