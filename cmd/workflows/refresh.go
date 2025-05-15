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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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

	if directory == "" {
		return fmt.Errorf("directory is required")
	}

	apiKey := viper.Get("api_key").(string)
	instanceURL := viper.Get("instance_url").(string)

	client := n8n.NewClient(instanceURL, apiKey)

	return RefreshWorkflowsWithClient(cmd, client, directory, dryRun, overwrite, output)
}

// RefreshWorkflowsWithClient is the testable version of RefreshWorkflows that accepts a client interface
func RefreshWorkflowsWithClient(cmd *cobra.Command, client n8n.ClientInterface, directory string, dryRun bool, overwrite bool, output string) error {

	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		if !dryRun {
			if err := os.MkdirAll(directory, 0755); err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
			cmd.Printf("Created directory: %s\n", directory)
		} else {
			cmd.Printf("Would create directory: %s\n", directory)
		}
	} else if err != nil {
		return fmt.Errorf("error accessing directory: %w", err)
	}

	workflowList, err := client.GetWorkflows()
	if err != nil {
		return fmt.Errorf("error fetching workflows: %w", err)
	}

	if workflowList == nil || workflowList.Data == nil || len(*workflowList.Data) == 0 {
		cmd.Println("No workflows found in n8n instance")
		return nil
	}

	localFiles := make(map[string]string)

	files, err := os.ReadDir(directory)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				ext := strings.ToLower(filepath.Ext(file.Name()))
				if ext == ".json" || ext == ".yaml" || ext == ".yml" {
					filePath := filepath.Join(directory, file.Name())
					if workflowID, err := ExtractWorkflowIDFromFile(filePath); err == nil && workflowID != "" {
						localFiles[workflowID] = filePath
					}
				}
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error reading directory: %w", err)
	}

	for _, workflow := range *workflowList.Data {
		if workflow.Id == nil || *workflow.Id == "" {
			cmd.Printf("Skipping workflow '%s' with no ID\n", workflow.Name)
			continue
		}

		sanitizedName := rootcmd.SanitizeFilename(workflow.Name)
		var filePath string
		var action string

		var extension string
		switch strings.ToLower(output) {
		case "yaml", "yml":
			extension = ".yaml"
		default:
			extension = ".json"
		}

		if existingPath, exists := localFiles[*workflow.Id]; exists && !overwrite {
			existingExt := filepath.Ext(existingPath)
			if (strings.ToLower(output) == "yaml" || strings.ToLower(output) == "yml") &&
				(strings.ToLower(existingExt) == ".json") {
				filePath = filepath.Join(directory, sanitizedName+extension)
				action = "Converting"
			} else if strings.ToLower(output) == "json" &&
				(strings.ToLower(existingExt) == ".yaml" || strings.ToLower(existingExt) == ".yml") {
				filePath = filepath.Join(directory, sanitizedName+extension)
				action = "Converting"
			} else {
				filePath = existingPath
				action = "Updating"
			}
		} else {
			filePath = filepath.Join(directory, sanitizedName+extension)
			action = "Creating"
		}

		var content []byte
		var err error

		switch strings.ToLower(filepath.Ext(filePath)) {
		case ".yaml", ".yml":
			var buf strings.Builder
			encoder := yaml.NewEncoder(&buf)
			encoder.SetIndent(2)
			if err := encoder.Encode(workflow); err != nil {
				return fmt.Errorf("error serializing workflow '%s' to YAML: %w", workflow.Name, err)
			}
			content = []byte(buf.String())
		default:
			content, err = json.MarshalIndent(workflow, "", "  ")
			if err != nil {
				return fmt.Errorf("error serializing workflow '%s' to JSON: %w", workflow.Name, err)
			}
		}

		needsUpdate := true
		if existingPath, exists := localFiles[*workflow.Id]; exists &&
			strings.EqualFold(filepath.Ext(existingPath), filepath.Ext(filePath)) {
			if _, fileErr := os.Stat(filePath); fileErr == nil {
				existingContent, readErr := os.ReadFile(filePath)
				if readErr == nil {
					ext := strings.ToLower(filepath.Ext(filePath))
					if ext == ".yaml" || ext == ".yml" {
						var existingWorkflow, newWorkflow n8n.Workflow
						if yamlErr := yaml.Unmarshal(existingContent, &existingWorkflow); yamlErr == nil {
							if yamlErr := yaml.Unmarshal(content, &newWorkflow); yamlErr == nil {
								existingJSON, _ := json.Marshal(existingWorkflow)
								newJSON, _ := json.Marshal(newWorkflow)
								needsUpdate = string(existingJSON) != string(newJSON)
							}
						}
					} else {
						needsUpdate = string(existingContent) != string(content)
					}
				}
			}
		}

		if !needsUpdate && action == "Updating" {
			cmd.Printf("No changes for workflow '%s' (ID: %s) in file: %s\n",
				workflow.Name, *workflow.Id, filePath)
			continue
		}

		if dryRun {
			if needsUpdate || action == "Creating" || action == "Converting" {
				cmd.Printf("Would %s workflow '%s' (ID: %s) to file: %s\n",
					strings.ToLower(action), workflow.Name, *workflow.Id, filePath)
			} else {
				cmd.Printf("No changes needed for workflow '%s' (ID: %s) in file: %s\n",
					workflow.Name, *workflow.Id, filePath)
			}
			continue
		}

		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("error writing workflow '%s' to file: %w", workflow.Name, err)
		}

		cmd.Printf("%s workflow '%s' (ID: %s) to file: %s\n",
			action, workflow.Name, *workflow.Id, filePath)
	}

	cmd.Println("Workflow refresh completed successfully")
	return nil
}
