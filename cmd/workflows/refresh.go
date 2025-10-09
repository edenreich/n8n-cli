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
	Long: `Refresh command fetches and updates the state of workflows in the directory from a specified n8n instance.
By default, only workflows that already exist in the directory will be refreshed. Use the --all flag to refresh all workflows.`,
	Args: cobra.ExactArgs(0),
	RunE: RefreshWorkflows,
}

func init() {
	refreshCmd.Flags().StringP("directory", "d", "", "Directory containing workflow files (JSON/YAML) (required)")
	refreshCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")
	refreshCmd.Flags().Bool("overwrite", false, "Overwrite existing files even if they have a different name")
	refreshCmd.Flags().StringP("output", "o", "json", "Output format for new workflow files (json or yaml)")
	refreshCmd.Flags().Bool("no-truncate", false, "Include all fields in output files, including null and optional fields")
	refreshCmd.Flags().Bool("all", false, "Refresh all workflows from n8n instance, not just those in the directory")
	refreshCmd.Flags().Bool("recursive", false, "Scan subdirectories recursively for workflow files")
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
	noTruncate, _ := cmd.Flags().GetBool("no-truncate")
	all, _ := cmd.Flags().GetBool("all")
	recursive, _ := cmd.Flags().GetBool("recursive")

	if directory == "" {
		return fmt.Errorf("directory is required")
	}

	apiKey := viper.Get("api_key").(string)
	instanceURL := viper.Get("instance_url").(string)

	client := n8n.NewClient(instanceURL, apiKey)

	minimal := !noTruncate

	return RefreshWorkflowsWithClient(cmd, client, directory, dryRun, overwrite, output, minimal, all, recursive)
}

// RefreshWorkflowsWithClient is the testable version of RefreshWorkflows that accepts a client interface
func RefreshWorkflowsWithClient(cmd *cobra.Command, client n8n.ClientInterface, directory string, dryRun bool, overwrite bool, output string, minimal bool, all bool, recursive bool) error {
	if err := ensureDirectoryExists(cmd, directory, dryRun); err != nil {
		return err
	}

	localFiles, err := extractLocalWorkflows(directory, recursive)
	if err != nil {
		return err
	}

	if all || len(localFiles) == 0 {
		cmd.Println("Refreshing all workflows from n8n instance")

		workflowList, err := client.GetWorkflows()
		if err != nil {
			return fmt.Errorf("error fetching workflows: %w", err)
		}

		if workflowList == nil || workflowList.Data == nil || len(*workflowList.Data) == 0 {
			cmd.Println("No workflows found in n8n instance")
			return nil
		}

		for _, workflow := range *workflowList.Data {
			if err := processWorkflow(cmd, workflow, localFiles, directory, dryRun, overwrite, output, minimal); err != nil {
				return err
			}
		}
	} else {
		cmd.Println("Refreshing only workflows that exist in the directory")

		refreshed := 0
		for workflowID := range localFiles {
			workflow, err := client.GetWorkflow(workflowID)
			if err != nil {
				cmd.Printf("Warning: Could not fetch workflow with ID %s: %v\n", workflowID, err)
				continue
			}

			if err := processWorkflow(cmd, *workflow, localFiles, directory, dryRun, overwrite, output, minimal); err != nil {
				return err
			}
			refreshed++
		}

		if refreshed == 0 {
			cmd.Println("No workflows were refreshed. Either the local workflows don't exist in the n8n instance or there was an error fetching them. Try refresh --all and delete the local files you don't want to track.")
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
func extractLocalWorkflows(directory string, recursive bool) (map[string]string, error) {
	localFiles := make(map[string]string)

	if recursive {
		err := filepath.WalkDir(directory, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}

			if d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(d.Name()))
			if ext != ".json" && ext != ".yaml" && ext != ".yml" {
				return nil
			}

			workflowID, err := ExtractWorkflowIDFromFile(path)
			if err != nil || workflowID == "" {
				return nil
			}

			if existingPath, exists := localFiles[workflowID]; exists {
				existingExt := strings.ToLower(filepath.Ext(existingPath))
				currentExt := strings.ToLower(filepath.Ext(path))

				if (currentExt == ".yaml" || currentExt == ".yml") && existingExt == ".json" {
					localFiles[workflowID] = path
				}
				return nil
			}

			localFiles[workflowID] = path
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking directory: %w", err)
		}
	} else {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, fmt.Errorf("error reading directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext != ".json" && ext != ".yaml" && ext != ".yml" {
				continue
			}

			path := filepath.Join(directory, entry.Name())
			workflowID, err := ExtractWorkflowIDFromFile(path)
			if err != nil || workflowID == "" {
				continue
			}

			if existingPath, exists := localFiles[workflowID]; exists {
				existingExt := strings.ToLower(filepath.Ext(existingPath))
				currentExt := strings.ToLower(filepath.Ext(path))

				if (currentExt == ".yaml" || currentExt == ".yml") && existingExt == ".json" {
					localFiles[workflowID] = path
				}
				continue
			}

			localFiles[workflowID] = path
		}
	}

	return localFiles, nil
}

// determineFilePathAndAction decides what file path and action to take for a workflow
func determineFilePathAndAction(workflow n8n.Workflow, localFiles map[string]string, directory string, output string, overwrite bool) (string, string) {
	sanitizedName := rootcmd.SanitizeFilename(workflow.Name)

	extension := ".json"

	if existingPath, exists := localFiles[*workflow.Id]; exists && output == "" {
		existingExt := strings.ToLower(filepath.Ext(existingPath))
		if existingExt == ".yaml" || existingExt == ".yml" {
			extension = existingExt
		}
	} else if strings.ToLower(output) == "yaml" || strings.ToLower(output) == "yml" {
		extension = ".yaml"
	}

	// Check if we have an existing file to preserve directory structure
	existingPath, exists := localFiles[*workflow.Id]
	if exists && !overwrite {
		// Preserve the existing file's directory structure
		existingExt := filepath.Ext(existingPath)
		if (strings.ToLower(output) == "yaml" || strings.ToLower(output) == "yml") && strings.ToLower(existingExt) == ".json" {
			// Convert to YAML but preserve directory structure
			dir := filepath.Dir(existingPath)
			convertedPath := filepath.Join(dir, sanitizedName+".yaml")
			return convertedPath, "Converting"
		}

		if strings.ToLower(output) == "json" && (strings.ToLower(existingExt) == ".yaml" || strings.ToLower(existingExt) == ".yml") {
			// Convert to JSON but preserve directory structure
			dir := filepath.Dir(existingPath)
			convertedPath := filepath.Join(dir, sanitizedName+".json")
			return convertedPath, "Converting"
		}

		return existingPath, "Updating"
	}

	// No existing file or overwrite mode - create in root directory
	defaultPath := filepath.Join(directory, sanitizedName+extension)
	return defaultPath, "Creating"
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

	// Ensure the directory exists before writing the file
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory '%s': %w", dir, err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("error writing workflow '%s' to file: %w", workflow.Name, err)
	}

	cmd.Printf("%s workflow '%s' (ID: %s) to file: %s\n",
		action, workflow.Name, *workflow.Id, filePath)

	return nil
}
