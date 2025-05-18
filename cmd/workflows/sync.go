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
	Short: "Synchronize workflows between local files and n8n instance",
	Long: `Synchronizes workflow files from a local directory to an n8n instance.

Examples:

  # Sync all workflow files from a directory
  n8n workflows sync --directory workflows/

  # Preview changes without applying them
  n8n workflows sync --directory workflows/ --dry-run

  # Sync and remove workflows that don't exist locally
  n8n workflows sync --directory workflows/ --prune

This command processes JSON and YAML workflow files and ensures they exist on your n8n instance:

1. Each workflow file is processed intelligently:
   - Workflows with IDs that exist on the server will be updated
   - Workflows with IDs that don't exist will be created
   - Workflows without IDs will be created as new
   - Active state (true/false) will be respected and applied

2. Common scenarios:
   - Development → Production: Create workflow files locally, test them, then sync to production
   - Backup: Store workflow configurations in a git repository for version control
   - Migration: Export workflows from one n8n instance and import to another
   - CI/CD: Automate workflow deployments in your delivery pipeline
   - Leverage AI-assisted development: Create workflows with Large Language Models (LLMs) and sync to n8n - streamlining workflow creation through code instead of manual UI interaction
   
3. File formats supported:
   - JSON: Standard n8n workflow export format
   - YAML: More readable alternative, ideal for version control

4. Additional examples:
   - Deploy workflows to production: 
     n8n workflows sync --directory workflows/production/

   - Migrate between environments (dev to staging):
     n8n workflows sync --directory workflows/dev/ --prune

   - Back up before a major change:
     mkdir -p backups/$(date +%Y%m%d) && \
     n8n workflows refresh --directory backups/$(date +%Y%m%d)/ --format json

   - In CI/CD pipelines:
     n8n workflows sync --directory workflows/ --dry-run && \
     n8n workflows sync --directory workflows/

5. Options:
   - Use --dry-run to preview changes without applying them
   - Use --prune to remove remote workflows that don't exist locally`,
	RunE: SyncWorkflows,
}

// TODO - imeplement tags updates during sync
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

	if prune {
		if err := PruneWorkflows(client, cmd, localWorkflowIDs); err != nil {
			return fmt.Errorf("error pruning workflows: %w", err)
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

	var w *n8n.Workflow
	var remoteWorkflow *n8n.Workflow

	if workflow.Id != nil && *workflow.Id != "" {
		remoteWorkflow, err = client.GetWorkflow(*workflow.Id)

		if err != nil {
			dryRunMsg := fmt.Sprintf("Would create workflow '%s' with ID %s from %s (ID specified but not found on server)", workflow.Name, *workflow.Id, filename)

			err = ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
				var err error
				w, err = client.CreateWorkflow(&workflow)
				if err != nil {
					return "", fmt.Errorf("error creating workflow: %w", err)
				}
				return fmt.Sprintf("Created workflow '%s' (ID: %s) from %s", w.Name, *w.Id, filename), nil
			})

			if err != nil {
				return err
			}
		} else {
			changes := DetectWorkflowChanges(&workflow, remoteWorkflow)

			if changes.NeedsUpdate {
				dryRunMsg := fmt.Sprintf("Would update workflow '%s' (ID: %s) from %s", workflow.Name, *workflow.Id, filename)

				err = ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
					var err error
					w, err = client.UpdateWorkflow(*workflow.Id, &workflow)
					if err != nil {
						return "", fmt.Errorf("error updating workflow: %w", err)
					}
					return fmt.Sprintf("Updated workflow '%s' (ID: %s) from %s", w.Name, *w.Id, filename), nil
				})

				if err != nil {
					return err
				}
			} else {
				w = remoteWorkflow
				if dryRun {
					cmd.Printf("No content changes for workflow '%s' (ID: %s) from %s\n", workflow.Name, *workflow.Id, filename)
				} else {
					cmd.Printf("No changes needed for workflow '%s' (ID: %s) from %s\n", workflow.Name, *workflow.Id, filename)
				}
			}
		}
	} else {
		dryRunMsg := fmt.Sprintf("Would create workflow '%s' from %s", workflow.Name, filename)

		err = ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
			var err error
			w, err = client.CreateWorkflow(&workflow)
			if err != nil {
				return "", fmt.Errorf("error creating workflow: %w", err)
			}
			return fmt.Sprintf("Created workflow '%s' (ID: %s) from %s", w.Name, *w.Id, filename), nil
		})

		if err != nil {
			return err
		}
	}

	var changes WorkflowChange
	if remoteWorkflow != nil {
		changes = DetectWorkflowChanges(&workflow, remoteWorkflow)
	} else {
		if workflow.Active != nil && *workflow.Active {
			changes.NeedsActivation = true
		} else if workflow.Active != nil && !*workflow.Active {
			changes.NeedsDeactivation = true
		}
	}

	if w != nil && workflow.Active != nil {
		workflowID := ""
		workflowName := workflow.Name

		if w.Id != nil {
			workflowID = *w.Id
		}

		if w.Name != "" {
			workflowName = w.Name
		}

		idInfo := ""
		if workflowID != "" {
			idInfo = fmt.Sprintf("(ID: %s)", workflowID)
		} else {
			idInfo = "(after creation)"
		}

		if *workflow.Active && changes.NeedsActivation {
			dryRunMsg := fmt.Sprintf("Would activate workflow '%s' %s", workflowName, idInfo)

			err = ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
				_, err := client.ActivateWorkflow(workflowID)
				if err != nil {
					return "", fmt.Errorf("error activating workflow: %w", err)
				}
				return fmt.Sprintf("Activated workflow '%s' %s", workflowName, idInfo), nil
			})

			if err != nil {
				return err
			}
		} else if !*workflow.Active && changes.NeedsDeactivation {
			dryRunMsg := fmt.Sprintf("Would deactivate workflow '%s' %s", workflowName, idInfo)

			err = ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
				_, err := client.DeactivateWorkflow(workflowID)
				if err != nil {
					return "", fmt.Errorf("error deactivating workflow: %w", err)
				}
				return fmt.Sprintf("Deactivated workflow '%s' %s", workflowName, idInfo), nil
			})

			if err != nil {
				return err
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

	dryRun := false
	if cmd.Flags().Changed("dry-run") {
		dryRun, _ = cmd.Flags().GetBool("dry-run")
	}

	for _, workflow := range *workflowList.Data {
		if workflow.Id == nil || *workflow.Id == "" {
			continue
		}

		if !localWorkflowIDs[*workflow.Id] {
			workflowID := *workflow.Id
			workflowName := workflow.Name

			dryRunMsg := fmt.Sprintf("Would delete workflow '%s' (ID: %s) that was not in local files", workflowName, workflowID)

			err := ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
				if err := client.DeleteWorkflow(workflowID); err != nil {
					return "", fmt.Errorf("error deleting workflow %s (%s): %w", workflowName, workflowID, err)
				}
				return fmt.Sprintf("Deleted workflow '%s' (ID: %s) that was not in local files", workflowName, workflowID), nil
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// WorkflowChange represents possible changes between local and remote workflows
type WorkflowChange struct {
	NeedsUpdate       bool
	NeedsActivation   bool
	NeedsDeactivation bool
}

// ExecuteOrDryRun is a helper function that either performs an action or shows what would happen
// based on whether dry run mode is enabled
func ExecuteOrDryRun(cmd *cobra.Command, dryRun bool, dryRunMsg string, fn func() (string, error)) error {
	if dryRun {
		cmd.Println(dryRunMsg)
		return nil
	}

	resultMsg, err := fn()
	if err != nil {
		return err
	}

	if resultMsg != "" {
		cmd.Println(resultMsg)
	}

	return nil
}

// DetectWorkflowChanges compares local and remote workflows to detect what changes are needed
func DetectWorkflowChanges(local *n8n.Workflow, remote *n8n.Workflow) WorkflowChange {
	changes := WorkflowChange{}

	if remote == nil {
		if local.Active != nil && *local.Active {
			changes.NeedsActivation = true
		}
		return changes
	}

	localCopy := *local
	localCopy.Id = nil
	localCopy.Active = nil
	localCopy.CreatedAt = nil
	localCopy.UpdatedAt = nil
	localCopy.Tags = nil

	remoteCopy := *remote
	remoteCopy.Id = nil
	remoteCopy.Active = nil
	remoteCopy.CreatedAt = nil
	remoteCopy.UpdatedAt = nil
	remoteCopy.Tags = nil

	localJSON, _ := json.Marshal(localCopy)
	remoteJSON, _ := json.Marshal(remoteCopy)

	changes.NeedsUpdate = string(localJSON) != string(remoteJSON)

	if local.Active != nil && remote.Active != nil {
		if *local.Active && !*remote.Active {
			changes.NeedsActivation = true
		} else if !*local.Active && *remote.Active {
			changes.NeedsDeactivation = true
		}
	} else if local.Active != nil && *local.Active {
		changes.NeedsActivation = true
	}

	return changes
}
