/*
Copyright © 2025 Eden Reich

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/edenreich/n8n-cli/config"
	"github.com/spf13/cobra"
)

// WorkflowData represents the structure of an n8n workflow
type WorkflowData struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// SyncConfig contains configuration for the sync operation
type SyncConfig struct {
	Directory   string
	APIBaseURL  string
	APIToken    string
	ActivateAll bool
	DryRun      bool
	Verbose     bool
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync JSON workflows to n8n instance",
	Long: `Sync command takes JSON workflow files from a specified directory 
and synchronizes them with your n8n instance. For example:

n8n-cli sync --directory hack/workflows

This will take all JSON workflow files from hack/workflows directory and 
upload them to the n8n instance specified in your environment configuration.

The command will:
1. Process all JSON files in the specified directory
2. For each workflow file:
   - If the workflow ID exists and is found on the n8n instance, it will update it
   - If the workflow doesn't exist or has no ID, it will create a new workflow
   - If the original workflow has "active": true, it will activate the workflow

Environment variables required:
- N8N_API_KEY: Your n8n API key
- N8N_INSTANCE_URL: URL of your n8n instance (e.g., https://your-instance.n8n.cloud)

These can be set in a .env file or as environment variables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		directory, _ := cmd.Flags().GetString("directory")
		activateAll, _ := cmd.Flags().GetBool("activate-all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if directory == "" {
			directory = "hack/workflows"
		}

		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("error loading configuration: %w", err)
		}

		fmt.Println("Starting workflow synchronization...")
		fmt.Printf("Using API URL: %s\n", cfg.APIBaseURL)
		fmt.Printf("Workflow directory: %s\n", directory)

		if dryRun {
			fmt.Println("DRY RUN MODE: No changes will be made to the n8n instance")
		}

		if activateAll {
			fmt.Println("All workflows will be activated after synchronization")
		}

		syncConfig := SyncConfig{
			Directory:   directory,
			APIBaseURL:  cfg.APIBaseURL,
			APIToken:    cfg.APIToken,
			ActivateAll: activateAll,
			DryRun:      dryRun,
			Verbose:     verbose,
		}

		if err = syncWorkflows(syncConfig); err != nil {
			return fmt.Errorf("error syncing workflows: %w", err)
		}

		fmt.Println("Workflow synchronization completed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringP("directory", "d", "hack/workflows", "Directory containing workflow JSON files")
	syncCmd.Flags().BoolP("activate-all", "a", false, "Activate all workflows after synchronization")
	syncCmd.Flags().BoolP("dry-run", "n", false, "Show what would be done without making changes")
	syncCmd.Flags().BoolP("verbose", "v", false, "Show detailed output during synchronization")
}

// syncWorkflows walks through the directory and syncs each workflow file
func syncWorkflows(config SyncConfig) error {
	_, err := os.Stat(config.Directory)
	if os.IsNotExist(err) {
		fmt.Printf("Directory %s does not exist, creating it...\n", config.Directory)
		if err := os.MkdirAll(config.Directory, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", config.Directory, err)
		}
		fmt.Printf("Created directory %s\n", config.Directory)
		return nil
	}

	serverWorkflows, err := getServerWorkflows(config.APIBaseURL, config.APIToken)
	if err != nil {
		return fmt.Errorf("failed to fetch server workflows: %v", err)
	}

	workflowNameToID := make(map[string]string)
	for _, workflow := range serverWorkflows {
		name, _ := workflow["name"].(string)
		id, _ := workflow["id"].(string)
		if name != "" && id != "" {
			workflowNameToID[name] = id
			if config.Verbose {
				fmt.Printf("Found server workflow: %s (ID: %s)\n", name, id)
			}
		}
	}

	return filepath.Walk(config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}

		workflowData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read workflow file %s: %v", path, err)
		}
		var workflow map[string]interface{}
		if err := json.Unmarshal(workflowData, &workflow); err != nil {
			return fmt.Errorf("failed to parse workflow JSON %s: %v", path, err)
		}

		workflowID, _ := workflow["id"].(string)
		workflowName, _ := workflow["name"].(string)

		if config.Verbose {
			fmt.Printf("Processing workflow: %s (ID: %s)\n", workflowName, workflowID)
		} else {
			fmt.Printf("Processing workflow: %s\n", workflowName)
		}

		serverID, existsOnServer := workflowNameToID[workflowName]
		if existsOnServer && config.Verbose {
			fmt.Printf("Found matching workflow on server with name: %s (ID: %s)\n", workflowName, serverID)
		}

		filteredWorkflow := map[string]interface{}{
			"name":        workflow["name"],
			"nodes":       workflow["nodes"],
			"connections": workflow["connections"],
			"settings":    workflow["settings"],
		}

		workflowData, err = json.Marshal(filteredWorkflow)
		if err != nil {
			return fmt.Errorf("failed to marshal filtered workflow: %v", err)
		}
		shouldActivate, _ := workflow["active"].(bool)

		if config.ActivateAll {
			shouldActivate = true
		}

		if existsOnServer && workflowID != serverID {
			if config.Verbose {
				fmt.Printf("Workflow with name '%s' exists on server with different ID: %s (local ID: %s)\n",
					workflowName, serverID, workflowID)
			}

			workflowID = serverID
		}

		if workflowID != "" {
			if config.DryRun {
				fmt.Printf("[DRY RUN] Would check if workflow %s exists\n", workflowName)
				fmt.Printf("[DRY RUN] Would update workflow: %s\n", workflowName)
				return nil
			}

			exists, err := checkWorkflowExists(workflowID, config.APIBaseURL, config.APIToken)
			if err != nil {
				return fmt.Errorf("error checking workflow existence: %v", err)
			}

			if exists {
				var workflowToUpdate map[string]interface{}
				if err := json.Unmarshal(workflowData, &workflowToUpdate); err != nil {
					return fmt.Errorf("failed to parse workflow data: %v", err)
				}
				workflowToUpdate["id"] = workflowID

				updateData, err := json.Marshal(workflowToUpdate)
				if err != nil {
					return fmt.Errorf("failed to marshal workflow with ID: %v", err)
				}

				err = updateWorkflow(workflowID, updateData, config.APIBaseURL, config.APIToken)
				if err != nil {
					return fmt.Errorf("error updating workflow %s: %v", workflowName, err)
				}
				fmt.Printf("Updated workflow: %s\n", workflowName)

				if shouldActivate {
					if err := activateWorkflow(workflowID, shouldActivate, config.APIBaseURL, config.APIToken); err != nil {
						fmt.Printf("Warning: Failed to activate workflow %s: %v\n", workflowName, err)
					} else {
						fmt.Printf("Activated workflow: %s\n", workflowName)
					}
				}

				return nil
			} else {
				serverID, nameMatchExists := workflowNameToID[workflowName]
				if nameMatchExists {
					fmt.Printf("Workflow with ID %s not found on server, but found workflow with name '%s' (server ID: %s)\n",
						workflowID, workflowName, serverID)

					var workflowToUpdate map[string]interface{}
					if err := json.Unmarshal(workflowData, &workflowToUpdate); err != nil {
						return fmt.Errorf("failed to parse workflow data: %v", err)
					}
					workflowToUpdate["id"] = serverID

					updateData, err := json.Marshal(workflowToUpdate)
					if err != nil {
						return fmt.Errorf("failed to marshal workflow with ID: %v", err)
					}

					err = updateWorkflow(serverID, updateData, config.APIBaseURL, config.APIToken)
					if err != nil {
						return fmt.Errorf("error updating workflow %s: %v", workflowName, err)
					}
					fmt.Printf("Updated existing workflow by name: %s (server ID: %s)\n", workflowName, serverID)

					if err := updateWorkflowFile(path, workflowID, serverID); err != nil {
						fmt.Printf("Warning: Failed to update workflow file with server ID: %v\n", err)
					} else {
						fmt.Printf("Updated workflow file with server ID: %s\n", path)
					}

					if shouldActivate {
						if err := activateWorkflow(serverID, shouldActivate, config.APIBaseURL, config.APIToken); err != nil {
							fmt.Printf("Warning: Failed to activate workflow %s: %v\n", workflowName, err)
						} else {
							fmt.Printf("Activated workflow: %s\n", workflowName)
						}
					}

					return nil
				} else {
					fmt.Printf("Workflow with ID %s not found on server, will create a new workflow\n", workflowID)

					newID, err := createWorkflow(workflowData, config.APIBaseURL, config.APIToken)
					if err != nil {
						return fmt.Errorf("error creating workflow %s: %v", workflowName, err)
					}
					fmt.Printf("Created new workflow: %s with ID: %s (original ID was: %s)\n", workflowName, newID, workflowID)

					if shouldActivate {
						if err := activateWorkflow(newID, shouldActivate, config.APIBaseURL, config.APIToken); err != nil {
							fmt.Printf("Warning: Failed to activate workflow %s: %v\n", workflowName, err)
						} else {
							fmt.Printf("Activated workflow: %s\n", workflowName)
						}
					}

					if err := updateWorkflowFile(path, workflowID, newID); err != nil {
						fmt.Printf("Warning: Failed to update workflow file with new ID: %v\n", err)
					} else {
						fmt.Printf("Updated workflow file with new ID: %s\n", path)
					}

					return nil
				}
			}
		}

		serverID, existsOnServer = workflowNameToID[workflowName]
		if existsOnServer {
			if config.DryRun {
				fmt.Printf("[DRY RUN] Would update existing workflow by name: %s (server ID: %s)\n", workflowName, serverID)
				return nil
			}

			var workflowToUpdate map[string]interface{}
			if err := json.Unmarshal(workflowData, &workflowToUpdate); err != nil {
				return fmt.Errorf("failed to parse workflow data: %v", err)
			}
			workflowToUpdate["id"] = serverID

			updateData, err := json.Marshal(workflowToUpdate)
			if err != nil {
				return fmt.Errorf("failed to marshal workflow with ID: %v", err)
			}

			err = updateWorkflow(serverID, updateData, config.APIBaseURL, config.APIToken)
			if err != nil {
				return fmt.Errorf("error updating workflow %s: %v", workflowName, err)
			}
			fmt.Printf("Updated existing workflow by name: %s (server ID: %s)\n", workflowName, serverID)

			if err := updateWorkflowFile(path, "", serverID); err != nil {
				fmt.Printf("Warning: Failed to update workflow file with server ID: %v\n", err)
			} else {
				fmt.Printf("Updated workflow file with server ID: %s\n", path)
			}

			if shouldActivate {
				if err := activateWorkflow(serverID, shouldActivate, config.APIBaseURL, config.APIToken); err != nil {
					fmt.Printf("Warning: Failed to activate workflow %s: %v\n", workflowName, err)
				} else {
					fmt.Printf("Activated workflow: %s\n", workflowName)
				}
			}

			return nil
		}

		if config.DryRun {
			fmt.Printf("[DRY RUN] Would create new workflow: %s\n", workflowName)
			return nil
		}

		newID, err := createWorkflow(workflowData, config.APIBaseURL, config.APIToken)
		if err != nil {
			return fmt.Errorf("error creating workflow %s: %v", workflowName, err)
		}

		fmt.Printf("Created new workflow: %s with ID: %s\n", workflowName, newID)

		if err := updateWorkflowFile(path, "", newID); err != nil {
			fmt.Printf("Warning: Failed to update workflow file with new ID: %v\n", err)
		} else {
			fmt.Printf("Updated workflow file with new ID: %s\n", path)
		}

		if shouldActivate {
			if err := activateWorkflow(newID, shouldActivate, config.APIBaseURL, config.APIToken); err != nil {
				fmt.Printf("Warning: Failed to activate workflow %s: %v\n", workflowName, err)
			} else {
				fmt.Printf("Activated workflow: %s\n", workflowName)
			}
		}

		return nil
	})
}

// checkWorkflowExists checks if a workflow exists on the n8n instance
func checkWorkflowExists(id, apiBaseURL, apiToken string) (bool, error) {
	url := fmt.Sprintf("%s/workflows/%s", apiBaseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("X-N8N-API-KEY", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	return resp.StatusCode == http.StatusOK, nil
}

// createWorkflow creates a new workflow on the n8n instance
func createWorkflow(data []byte, apiBaseURL, apiToken string) (string, error) {
	url := fmt.Sprintf("%s/workflows", apiBaseURL)

	var workflow map[string]interface{}
	if err := json.Unmarshal(data, &workflow); err != nil {
		return "", fmt.Errorf("failed to parse workflow data: %v", err)
	}

	delete(workflow, "id")

	dataWithoutID, err := json.Marshal(workflow)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow without ID: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataWithoutID))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	id, ok := response["id"].(string)
	if !ok {
		return "", fmt.Errorf("could not extract workflow ID from response")
	}

	return id, nil
}

// updateWorkflow updates an existing workflow on the n8n instance and returns the latest workflow data
func updateWorkflow(id string, data []byte, apiBaseURL, apiToken string) error {
	url := fmt.Sprintf("%s/workflows/%s", apiBaseURL, id)

	var workflowData map[string]interface{}
	if err := json.Unmarshal(data, &workflowData); err != nil {
		return fmt.Errorf("failed to parse workflow data for update: %v", err)
	}

	delete(workflowData, "id")

	dataWithoutID, err := json.Marshal(workflowData)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow data without ID: %v", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(dataWithoutID))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	latestWorkflow, err := fetchWorkflow(id, apiBaseURL, apiToken)
	if err != nil {
		return fmt.Errorf("failed to fetch updated workflow data: %v", err)
	}

	workflowData = latestWorkflow

	return nil
}

// activateWorkflow activates a workflow if needed
func activateWorkflow(id string, shouldActivate bool, apiBaseURL, apiToken string) error {
	if !shouldActivate {
		return nil
	}

	url := fmt.Sprintf("%s/workflows/%s/activate", apiBaseURL, id)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// updateWorkflowFile updates the workflow file with the new ID
func updateWorkflowFile(path, oldID, newID string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read workflow file %s: %v", path, err)
	}

	var workflow map[string]interface{}
	if err := json.Unmarshal(data, &workflow); err != nil {
		return fmt.Errorf("failed to parse workflow JSON %s: %v", path, err)
	}

	workflow["id"] = newID

	updatedData, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow data: %v", err)
	}

	if err := os.WriteFile(path, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated workflow to file %s: %v", path, err)
	}

	return nil
}

// fetchWorkflow gets the latest workflow data from the n8n instance
func fetchWorkflow(id, apiBaseURL, apiToken string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/workflows/%s", apiBaseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var workflow map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow data: %v", err)
	}

	return workflow, nil
}
