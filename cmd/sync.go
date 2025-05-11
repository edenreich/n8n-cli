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

	"github.com/joho/godotenv"
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

These can be set in a .env file or as environment variables.`, Run: func(cmd *cobra.Command, args []string) {
		directory, _ := cmd.Flags().GetString("directory")
		activateAll, _ := cmd.Flags().GetBool("activate-all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if directory == "" {
			directory = "hack/workflows"
		}

		err := loadEnv()
		if err != nil {
			fmt.Println("Error loading environment variables:", err)
			return
		}

		apiToken := os.Getenv("N8N_API_KEY")
		instanceURL := os.Getenv("N8N_INSTANCE_URL")

		if apiToken == "" || instanceURL == "" {
			fmt.Println("N8N_API_KEY and N8N_INSTANCE_URL environment variables must be set")
			fmt.Println("You can create a .env file based on .env.example")
			return
		}

		apiBaseURL := ensureAPIPrefix(instanceURL)

		fmt.Println("Starting workflow synchronization...")
		fmt.Printf("Using API URL: %s\n", apiBaseURL)
		fmt.Printf("Workflow directory: %s\n", directory)

		if dryRun {
			fmt.Println("DRY RUN MODE: No changes will be made to the n8n instance")
		}

		if activateAll {
			fmt.Println("All workflows will be activated after synchronization")
		}

		config := SyncConfig{
			Directory:   directory,
			APIBaseURL:  apiBaseURL,
			APIToken:    apiToken,
			ActivateAll: activateAll,
			DryRun:      dryRun,
			Verbose:     verbose,
		}

		err = syncWorkflows(config)
		if err != nil {
			fmt.Println("Error syncing workflows:", err)
			return
		}

		fmt.Println("Workflow synchronization completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringP("directory", "d", "hack/workflows", "Directory containing workflow JSON files")
	syncCmd.Flags().BoolP("activate-all", "a", false, "Activate all workflows after synchronization")
	syncCmd.Flags().BoolP("dry-run", "n", false, "Show what would be done without making changes")
	syncCmd.Flags().BoolP("verbose", "v", false, "Show detailed output during synchronization")
}

// loadEnv loads environment variables from .env file
func loadEnv() error {
	_ = godotenv.Load()
	return nil
}

// ensureAPIPrefix ensures the URL has the /api/v1 prefix
func ensureAPIPrefix(url string) string {
	url = strings.TrimSuffix(url, "/")

	if !strings.HasSuffix(url, "/api/v1") {
		return url + "/api/v1"
	}

	return url
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

		if workflowID != "" {
			if config.DryRun {
				fmt.Printf("[DRY RUN] Would check if workflow %s exists\n", workflowName)
				exists := false
				if !exists {
					fmt.Printf("[DRY RUN] Would create workflow: %s\n", workflowName)
					return nil
				}
			}

			exists, err := checkWorkflowExists(workflowID, config.APIBaseURL, config.APIToken)
			if err != nil {
				return fmt.Errorf("error checking workflow existence: %v", err)
			}

			if exists {
				if config.DryRun {
					fmt.Printf("[DRY RUN] Would update workflow: %s\n", workflowName)
				} else {
					err := updateWorkflow(workflowID, workflowData, config.APIBaseURL, config.APIToken)
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
				}

				return nil
			}
		}

		if config.DryRun {
			fmt.Printf("[DRY RUN] Would create workflow: %s\n", workflowName)
			return nil
		}

		newID, err := createWorkflow(workflowData, config.APIBaseURL, config.APIToken)
		if err != nil {
			return fmt.Errorf("error creating workflow %s: %v", workflowName, err)
		}

		fmt.Printf("Created workflow: %s with ID: %s\n", workflowName, newID)

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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
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

// updateWorkflow updates an existing workflow on the n8n instance
func updateWorkflow(id string, data []byte, apiBaseURL, apiToken string) error {
	url := fmt.Sprintf("%s/workflows/%s", apiBaseURL, id)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
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
