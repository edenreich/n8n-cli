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
package workflows

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the state of workflows in the directory from n8n instance",
	Long:  `Refresh command fetches and updates the state of workflows in the directory from a specified n8n instance.`,
	Args:  cobra.ExactArgs(0),
	RunE:  RefreshWorkflows,
}

func init() {
	refreshCmd.Flags().StringP("directory", "d", "", "Directory containing workflow files (JSON/YAML) (required)")
	refreshCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")
	refreshCmd.Flags().Bool("overwrite", false, "Overwrite existing files even if they have a different name")
	refreshCmd.Flags().StringP("output", "o", "json", "Output format for new workflow files (json or yaml)")
	refreshCmd.Flags().Bool("minimal", true, "Minimize workflow files by removing null and optional fields")
	rootcmd.GetWorkflowsCmd().AddCommand(refreshCmd)

	// nolint:errcheck
	refreshCmd.MarkFlagRequired("directory")
}

// RefreshWorkflows refreshes workflow files from n8n instance
func RefreshWorkflows(cmd *cobra.Command, args []string) error {
	cmd.Println("Refreshing workflows...")
	directory, _ := cmd.Flags().GetString("directory")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	output, _ := cmd.Flags().GetString("output")
	minimal, _ := cmd.Flags().GetBool("minimal")

	if directory == "" {
		return fmt.Errorf("directory is required")
	}

	apiKey := viper.Get("api_key").(string)
	instanceURL := viper.Get("instance_url").(string)

	client := n8n.NewClient(instanceURL, apiKey)

	return RefreshWorkflowsWithClient(cmd, client, directory, dryRun, overwrite, output, minimal)
}

// RefreshWorkflowsWithClient is the testable version of RefreshWorkflows that accepts a client interface
func RefreshWorkflowsWithClient(cmd *cobra.Command, client n8n.ClientInterface, directory string, dryRun bool, overwrite bool, output string, minimal bool) error {
	if err := ensureDirectoryExists(cmd, directory, dryRun); err != nil {
		return err
	}

	workflowList, err := client.GetWorkflows()
	if err != nil {
		return fmt.Errorf("error fetching workflows: %w", err)
	}

	if workflowList == nil || workflowList.Data == nil || len(*workflowList.Data) == 0 {
		cmd.Println("No workflows found in n8n instance")
		return nil
	}

	localFiles, err := extractLocalWorkflows(directory)
	if err != nil {
		return err
	}

	for _, workflow := range *workflowList.Data {
		if err := processWorkflow(cmd, workflow, localFiles, directory, dryRun, overwrite, output, minimal); err != nil {
			return err
		}
	}

	cmd.Println("Workflow refresh completed successfully")
	return nil
}

// ensureDirectoryExists checks if the directory exists and creates it if needed
func ensureDirectoryExists(cmd *cobra.Command, directory string, dryRun bool) error {
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		if dryRun {
			cmd.Printf("Would create directory: %s\n", directory)
			return nil
		}

		if err := os.MkdirAll(directory, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		cmd.Printf("Created directory: %s\n", directory)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error accessing directory: %w", err)
	}

	return nil
}

// extractLocalWorkflows reads local workflow files and returns a map of workflow IDs to file paths
func extractLocalWorkflows(directory string) (map[string]string, error) {
	localFiles := make(map[string]string)

	files, err := os.ReadDir(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return localFiles, nil
		}
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			continue
		}

		filePath := filepath.Join(directory, file.Name())
		workflowID, err := ExtractWorkflowIDFromFile(filePath)
		if err != nil || workflowID == "" {
			continue
		}

		localFiles[workflowID] = filePath
	}

	return localFiles, nil
}

// determineFilePathAndAction decides what file path and action to take for a workflow
func determineFilePathAndAction(workflow n8n.Workflow, localFiles map[string]string, directory string, output string, overwrite bool) (string, string) {
	sanitizedName := rootcmd.SanitizeFilename(workflow.Name)
	extension := ".json"
	if strings.ToLower(output) == "yaml" || strings.ToLower(output) == "yml" {
		extension = ".yaml"
	}

	defaultPath := filepath.Join(directory, sanitizedName+extension)

	existingPath, exists := localFiles[*workflow.Id]
	if !exists || overwrite {
		return defaultPath, "Creating"
	}

	existingExt := filepath.Ext(existingPath)
	if (strings.ToLower(output) == "yaml" || strings.ToLower(output) == "yml") && strings.ToLower(existingExt) == ".json" {
		return defaultPath, "Converting"
	}

	if strings.ToLower(output) == "json" && (strings.ToLower(existingExt) == ".yaml" || strings.ToLower(existingExt) == ".yml") {
		return defaultPath, "Converting"
	}

	return existingPath, "Updating"
}

// serializeWorkflow serializes a workflow to JSON or YAML
func serializeWorkflow(workflow n8n.Workflow, filePath string, minimal bool) ([]byte, error) {
	encoder := n8n.NewWorkflowEncoder(minimal)

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".yaml" || ext == ".yml" {
		yamlData, err := encoder.EncodeToYAML(workflow)
		if err != nil {
			return nil, fmt.Errorf("error serializing workflow '%s' to YAML: %w", workflow.Name, err)
		}

		return yamlData, nil
	}

	jsonData, err := encoder.EncodeToJSON(workflow)
	if err != nil {
		return nil, fmt.Errorf("error serializing workflow '%s' to JSON: %w", workflow.Name, err)
	}

	return jsonData, nil
}

// workflowNeedsUpdate compares existing workflow file content with new content
func workflowNeedsUpdate(filePath string, existingPath string, content []byte, minimal bool) bool {
	if _, fileErr := os.Stat(filePath); fileErr != nil {
		return true
	}

	if !strings.EqualFold(filepath.Ext(existingPath), filepath.Ext(filePath)) {
		return true
	}

	existingContent, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return true
	}

	decoder := n8n.NewWorkflowDecoder()

	existingWorkflow, err := decoder.DecodeFromBytes(existingContent)
	if err != nil {
		return true
	}

	newWorkflow, err := decoder.DecodeFromBytes(content)
	if err != nil {
		return true
	}

	return rootcmd.DetectWorkflowDrift(existingWorkflow, newWorkflow, minimal)
}

// Both compareYAMLContent and compareJSONContent functions have been replaced
// by the DetectWorkflowDrift utility function in the cmd package

// processWorkflow handles processing of a single workflow
func processWorkflow(cmd *cobra.Command, workflow n8n.Workflow, localFiles map[string]string,
	directory string, dryRun bool, overwrite bool, output string, minimal bool) error {

	if workflow.Id == nil || *workflow.Id == "" {
		cmd.Printf("Skipping workflow '%s' with no ID\n", workflow.Name)
		return nil
	}

	filePath, action := determineFilePathAndAction(workflow, localFiles, directory, output, overwrite)
	existingPath := localFiles[*workflow.Id]

	content, err := serializeWorkflow(workflow, filePath, minimal)
	if err != nil {
		return err
	}

	needsUpdate := true
	if action == "Updating" {
		needsUpdate = workflowNeedsUpdate(filePath, existingPath, content, minimal)
		if !needsUpdate {
			cmd.Printf("No changes for workflow '%s' (ID: %s) in file: %s\n",
				workflow.Name, *workflow.Id, filePath)
			return nil
		}
	}

	if dryRun {
		if needsUpdate || action == "Creating" || action == "Converting" {
			cmd.Printf("Would %s workflow '%s' (ID: %s) to file: %s\n",
				strings.ToLower(action), workflow.Name, *workflow.Id, filePath)
		} else {
			cmd.Printf("No changes needed for workflow '%s' (ID: %s) in file: %s\n",
				workflow.Name, *workflow.Id, filePath)
		}
		return nil
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("error writing workflow '%s' to file: %w", workflow.Name, err)
	}

	cmd.Printf("%s workflow '%s' (ID: %s) to file: %s\n",
		action, workflow.Name, *workflow.Id, filePath)

	return nil
}
