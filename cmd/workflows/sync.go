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

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// syncCmd represents the sync command
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync workflows to n8n instance",
	Long: `Sync command takes workflow files (JSON or YAML format) from a specified directory 
and synchronizes them with your n8n instance. For example:

n8n-cli workflows sync --directory workflows/

This will take all JSON and YAML workflow files from workflows directory and 
upload them to the n8n instance specified in your environment configuration.

The command will:
1. Process all JSON and YAML files in the specified directory
2. For each workflow file:
   - If the workflow ID exists and is found on the n8n instance, it will update it
   - If the workflow doesn't exist or has no ID, it will create a new workflow
   - If the workflow has "active": true in its definition, it will be activated automatically
3. If the --dry-run flag is set, it will show what would be done without making any changes
4. If the --prune flag is set, it will remove workflows that are not present in the directory`,
	RunE: SyncWorkflows,
}

func init() {
	cmd.GetWorkflowsCmd().AddCommand(SyncCmd)

	SyncCmd.Flags().StringP("directory", "d", "", "Directory containing workflow files (JSON/YAML) (required)")
	SyncCmd.Flags().Bool("dry-run", false, "Show what would be uploaded without making changes")
	SyncCmd.Flags().Bool("prune", false, "Remove workflows that are not present in the directory")

	// nolint:errcheck
	SyncCmd.MarkFlagRequired("directory")
}

// SyncWorkflows syncs workflow files from a directory to n8n
func SyncWorkflows(cmd *cobra.Command, args []string) error {
	cmd.Println("Syncing workflows...")
	directory, _ := cmd.Flags().GetString("directory")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	prune, _ := cmd.Flags().GetBool("prune")

	if directory == "" {
		return fmt.Errorf("directory is required")
	}

	files, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	apiKey := viper.Get("api_key").(string)
	instanceURL := viper.Get("instance_url").(string)

	client := n8n.NewClient(instanceURL, apiKey)

	localWorkflowIDs := make(map[string]bool)

	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".json" || ext == ".yaml" || ext == ".yml" {
				filePath := filepath.Join(directory, file.Name())

				if workflowID, err := ExtractWorkflowIDFromFile(filePath); err == nil && workflowID != "" {
					localWorkflowIDs[workflowID] = true
				}

				if err = ProcessWorkflowFile(client, cmd, filePath, dryRun, prune); err != nil {
					return fmt.Errorf("error processing workflow file %s: %w", filePath, err)
				}
			}
		}
	}

	if prune && !dryRun {
		if err := PruneWorkflows(client, cmd, localWorkflowIDs); err != nil {
			return fmt.Errorf("error pruning workflows: %w", err)
		}
	} else if prune && dryRun {
		if err := DryRunPruneWorkflows(client, cmd, localWorkflowIDs); err != nil {
			return fmt.Errorf("error during dry run prune: %w", err)
		}
	}

	return nil
}

// ProcessWorkflowFile processes a workflow file and uploads it to n8n
func ProcessWorkflowFile(client n8n.ClientInterface, cmd *cobra.Command, filePath string, dryRun bool, prune bool) error {
	var workflow n8n.Workflow
	var err error

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	filename := filepath.Base(filePath)

	switch ext {
	case ".json":
		if err = json.Unmarshal(content, &workflow); err != nil {
			return fmt.Errorf("error parsing JSON workflow: %w", err)
		}
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(content, &workflow); err != nil {
			return fmt.Errorf("error parsing YAML workflow: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	var result *n8n.Workflow
	if workflow.Id != nil && *workflow.Id != "" {
		if dryRun {
			cmd.Printf("Would update workflow '%s' (ID: %s) from %s\n", workflow.Name, *workflow.Id, filename)
		} else {
			result, err = client.UpdateWorkflow(*workflow.Id, &workflow)
			if err != nil {
				return fmt.Errorf("error updating workflow: %w", err)
			}
			cmd.Printf("Updated workflow '%s' (ID: %s) from %s\n", result.Name, *result.Id, filename)
		}
	} else {
		if dryRun {
			cmd.Printf("Would create workflow '%s' from %s\n", workflow.Name, filename)
		} else {
			result, err = client.CreateWorkflow(&workflow)
			if err != nil {
				return fmt.Errorf("error creating workflow: %w", err)
			}
			cmd.Printf("Created workflow '%s' (ID: %s) from %s\n", result.Name, *result.Id, filename)
		}
	}

	if !dryRun && result != nil {
		if workflow.Active != nil && *workflow.Active {
			_, err = client.ActivateWorkflow(*result.Id)
			if err != nil {
				return fmt.Errorf("error activating workflow: %w", err)
			}
			cmd.Printf("Activated workflow '%s' (ID: %s)\n", result.Name, *result.Id)
		}
	} else if dryRun {
		if workflow.Active != nil && *workflow.Active {
			if workflow.Id != nil && *workflow.Id != "" {
				cmd.Printf("Would activate workflow '%s' (ID: %s)\n", workflow.Name, *workflow.Id)
			} else {
				cmd.Printf("Would activate workflow '%s' (after creation)\n", workflow.Name)
			}
		}
	}

	return nil
}

// ExtractWorkflowIDFromFile reads a workflow file and extracts the workflow ID if present
func ExtractWorkflowIDFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".json":
		var workflow n8n.Workflow
		if err = json.Unmarshal(content, &workflow); err != nil {
			return "", fmt.Errorf("error parsing JSON workflow: %w", err)
		}

		if workflow.Id != nil {
			return *workflow.Id, nil
		}
	case ".yaml", ".yml":
		var workflowMap map[string]interface{}
		if err = yaml.Unmarshal(content, &workflowMap); err != nil {
			return "", fmt.Errorf("error parsing YAML workflow: %w", err)
		}

		if id, ok := workflowMap["id"]; ok {
			if idStr, ok := id.(string); ok {
				return idStr, nil
			}
		}
	}

	return "", nil
}

// PruneWorkflows removes workflows from n8n that are not in the local workflow files
func PruneWorkflows(client n8n.ClientInterface, cmd *cobra.Command, localWorkflowIDs map[string]bool) error {
	workflowList, err := client.GetWorkflows()
	if err != nil {
		return fmt.Errorf("error getting workflows from n8n: %w", err)
	}

	if workflowList == nil || workflowList.Data == nil {
		return fmt.Errorf("no workflows found in n8n instance")
	}

	for _, workflow := range *workflowList.Data {
		if workflow.Id == nil || *workflow.Id == "" {
			continue
		}

		if !localWorkflowIDs[*workflow.Id] {
			if err := client.DeleteWorkflow(*workflow.Id); err != nil {
				return fmt.Errorf("error deleting workflow %s (%s): %w", workflow.Name, *workflow.Id, err)
			}
			cmd.Printf("Deleted workflow '%s' (ID: %s) that was not in local files\n", workflow.Name, *workflow.Id)
		}
	}

	return nil
}

// DryRunPruneWorkflows simulates pruning workflows without actually deleting anything
func DryRunPruneWorkflows(client n8n.ClientInterface, cmd *cobra.Command, localWorkflowIDs map[string]bool) error {
	workflowList, err := client.GetWorkflows()
	if err != nil {
		return fmt.Errorf("error getting workflows from n8n: %w", err)
	}

	if workflowList == nil || workflowList.Data == nil {
		return fmt.Errorf("no workflows found in n8n instance")
	}

	for _, workflow := range *workflowList.Data {
		if workflow.Id == nil || *workflow.Id == "" {
			continue
		}

		if !localWorkflowIDs[*workflow.Id] {
			cmd.Printf("Would delete workflow '%s' (ID: %s) that was not in local files\n", workflow.Name, *workflow.Id)
		}
	}

	return nil
}
