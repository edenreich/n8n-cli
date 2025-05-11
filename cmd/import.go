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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// ImportConfig contains configuration for the import operation
type ImportConfig struct {
	Directory  string
	APIBaseURL string
	APIToken   string
	Verbose    bool
	DryRun     bool
	All        bool
	WorkflowID string
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import workflows from an n8n instance",
	Long: `Import command fetches workflows from your n8n instance and saves them as JSON files
in a specified directory. For example:

n8n-cli import --directory hack/workflows

This will fetch all workflows from your n8n instance and save them in the hack/workflows directory.

You can also import a specific workflow by ID:

n8n-cli import --workflow-id 123 --directory hack/workflows

Environment variables required:
- N8N_API_KEY: Your n8n API key
- N8N_INSTANCE_URL: URL of your n8n instance (e.g., https://your-instance.n8n.cloud)

These can be set in a .env file or as environment variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		directory, _ := cmd.Flags().GetString("directory")
		workflowID, _ := cmd.Flags().GetString("workflow-id")
		all, _ := cmd.Flags().GetBool("all")
		verbose, _ := cmd.Flags().GetBool("verbose")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if directory == "" {
			directory = "hack/workflows"
		}

		if !all && workflowID == "" {
			all = true
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

		fmt.Println("Starting workflow import...")
		fmt.Printf("Using API URL: %s\n", apiBaseURL)
		fmt.Printf("Workflow directory: %s\n", directory)

		if dryRun {
			fmt.Println("DRY RUN MODE: No files will be created")
		}

		config := ImportConfig{
			Directory:  directory,
			APIBaseURL: apiBaseURL,
			APIToken:   apiToken,
			DryRun:     dryRun,
			Verbose:    verbose,
			All:        all,
			WorkflowID: workflowID,
		}

		err = importWorkflows(config)
		if err != nil {
			fmt.Println("Error importing workflows:", err)
			return
		}

		fmt.Println("Workflow import completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("directory", "d", "hack/workflows", "Directory to save workflow JSON files")
	importCmd.Flags().StringP("workflow-id", "w", "", "ID of a specific workflow to import")
	importCmd.Flags().BoolP("all", "a", false, "Import all workflows")
	importCmd.Flags().BoolP("dry-run", "n", false, "Show what would be done without making changes")
	importCmd.Flags().BoolP("verbose", "v", false, "Show detailed output during import")
}

// importWorkflows fetches workflows from the n8n instance and saves them to the directory
func importWorkflows(config ImportConfig) error {
	if _, err := os.Stat(config.Directory); os.IsNotExist(err) {
		if config.Verbose {
			fmt.Printf("Directory %s does not exist, creating it...\n", config.Directory)
		}
		if !config.DryRun {
			if err := os.MkdirAll(config.Directory, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", config.Directory, err)
			}
		}
		if config.Verbose {
			fmt.Printf("Created directory %s\n", config.Directory)
		}
	}

	if config.All {
		return importAllWorkflows(config)
	} else {
		return importWorkflowByID(config)
	}
}

// importAllWorkflows imports all workflows from the n8n instance
func importAllWorkflows(config ImportConfig) error {
	url := fmt.Sprintf("%s/workflows", config.APIBaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-N8N-API-KEY", config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse workflow response: %v", err)
	}

	for _, workflow := range response.Data {
		workflowID, _ := workflow["id"].(string)
		workflowName, _ := workflow["name"].(string)

		if config.Verbose {
			fmt.Printf("Found workflow: %s (ID: %s)\n", workflowName, workflowID)
		} else {
			fmt.Printf("Found workflow: %s\n", workflowName)
		}

		// Get the full workflow data with all details
		err := importWorkflowByIDWithConfig(workflowID, config)
		if err != nil {
			fmt.Printf("Error importing workflow %s: %v\n", workflowID, err)
			continue
		}
	}

	return nil
}

// importWorkflowByID imports a specific workflow by ID
func importWorkflowByID(config ImportConfig) error {
	return importWorkflowByIDWithConfig(config.WorkflowID, config)
}

// importWorkflowByIDWithConfig imports a specific workflow by ID with the given config
func importWorkflowByIDWithConfig(workflowID string, config ImportConfig) error {
	url := fmt.Sprintf("%s/workflows/%s", config.APIBaseURL, workflowID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-N8N-API-KEY", config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var workflow map[string]interface{}
	if err := json.Unmarshal(body, &workflow); err != nil {
		return fmt.Errorf("failed to parse workflow response: %v", err)
	}

	var workflowData map[string]interface{}
	if data, ok := workflow["data"].(map[string]interface{}); ok {
		workflowData = data
	} else {
		workflowData = workflow
	}

	workflowName, _ := workflowData["name"].(string)
	if workflowName == "" {
		workflowName = fmt.Sprintf("workflow-%s", workflowID)
	}

	filename := sanitizeFilename(workflowName) + ".json"
	filePath := filepath.Join(config.Directory, filename)

	if config.Verbose {
		fmt.Printf("Importing workflow: %s (ID: %s) to %s\n", workflowName, workflowID, filePath)
	} else {
		fmt.Printf("Importing workflow: %s to %s\n", workflowName, filePath)
	}

	if !config.DryRun {
		prettyJSON, err := json.MarshalIndent(workflowData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format workflow JSON: %v", err)
		}

		prettyJSON = append(prettyJSON, '\n')

		err = os.WriteFile(filePath, prettyJSON, 0644)
		if err != nil {
			return fmt.Errorf("failed to write workflow file %s: %v", filePath, err)
		}
	}

	return nil
}

// sanitizeFilename converts a workflow name to a valid filename
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	return name
}
