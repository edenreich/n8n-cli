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
	"reflect"
	"strings"

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/logger"
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

  # Sync without refreshing local files afterward
  n8n workflows sync --directory workflows/ --refresh=false

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
   - Use --prune to remove remote workflows that don't exist locally
   - Use --refresh=false to prevent refreshing local files with remote state after sync
   - Use --output to specify the format (json or yaml) for refreshed workflow files
   - Use --all to refresh all workflows from n8n instance, not just those in the directory`,
	RunE: SyncWorkflows,
}

func init() {
	cmd.GetWorkflowsCmd().AddCommand(SyncCmd)

	SyncCmd.Flags().StringP("directory", "d", "", "Directory containing workflow files (JSON/YAML) (required)")
	SyncCmd.Flags().Bool("dry-run", false, "Show what would be uploaded without making changes")
	SyncCmd.Flags().Bool("prune", false, "Remove workflows that are not present in the directory")
	SyncCmd.Flags().Bool("refresh", true, "Refresh the local state with the remote state")
	SyncCmd.Flags().StringP("output", "o", "", "Output format for refreshed workflow files (json or yaml). If not specified, uses the existing file extension in the directory")
	SyncCmd.Flags().Bool("all", false, "Refresh all workflows from n8n instance when refreshing, not just those in the directory")

	// nolint:errcheck
	SyncCmd.MarkFlagRequired("directory")
}

// SyncWorkflows syncs workflow files from a directory to n8n
func SyncWorkflows(cmd *cobra.Command, args []string) error {
	cmd.Println("Syncing workflows...")
	directory, _ := cmd.Flags().GetString("directory")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	prune, _ := cmd.Flags().GetBool("prune")
	refresh, _ := cmd.Flags().GetBool("refresh")
	all, _ := cmd.Flags().GetBool("all")

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
	updatedWorkflows := make(map[string]bool)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".json" || ext == ".yaml" || ext == ".yml" {
			filePath := filepath.Join(directory, file.Name())

			if workflowID, err := ExtractWorkflowIDFromFile(filePath); err == nil && workflowID != "" {
				localWorkflowIDs[workflowID] = true
			}

			result, err := ProcessWorkflowFile(client, cmd, filePath, dryRun, prune)
			if err != nil {
				return fmt.Errorf("error processing workflow file %s: %w", filePath, err)
			}

			if result.WorkflowID != "" {
				updatedWorkflows[result.WorkflowID] = true
			}
		}
	}

	if prune {
		if err := PruneWorkflows(client, cmd, localWorkflowIDs); err != nil {
			return fmt.Errorf("error pruning workflows: %w", err)
		}
	}

	if refresh && !dryRun && len(updatedWorkflows) > 0 {
		cmd.Println("Refreshing local workflow files with remote state...")

		noTruncate := false
		minimal := !noTruncate
		overwrite := true
		output, _ := cmd.Flags().GetString("output")

		if output == "" {
			cmd.Println("No output format specified, maintaining existing file formats")
		}

		if err := RefreshWorkflowsWithClient(cmd, client, directory, false, overwrite, output, minimal, all); err != nil {
			return fmt.Errorf("error refreshing workflows after sync: %w", err)
		}

		cmd.Println("Local workflow files updated successfully with remote state")
	}

	return nil
}

// WorkflowResult contains the result of processing a workflow file
type WorkflowResult struct {
	WorkflowID string
	Name       string
	FilePath   string
	Created    bool
	Updated    bool
}

// ProcessWorkflowFile processes a workflow file and uploads it to n8n
func ProcessWorkflowFile(client n8n.ClientInterface, cmd *cobra.Command, filePath string, dryRun bool, prune bool) (WorkflowResult, error) {
	var workflow n8n.Workflow
	var err error
	result := WorkflowResult{
		FilePath: filePath,
	}

	logger.Debug("Processing file: %s", filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return result, fmt.Errorf("error reading file: %w", err)
	}

	logger.Debug("File size: %d bytes", len(content))
	if len(content) > 0 {
		preview := string(content)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		logger.Debug("Content preview: %s", preview)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	filename := filepath.Base(filePath)

	switch ext {
	case ".json":
		logger.Debug("Parsing as JSON: %s", filename)
		if err = json.Unmarshal(content, &workflow); err != nil {
			logger.Debug("JSON parsing error: %v", err)
			return result, fmt.Errorf("error parsing JSON workflow: %w", err)
		}
	case ".yaml", ".yml":
		logger.Debug("Parsing as YAML: %s", filename)

		decoder := n8n.NewWorkflowDecoder()
		workflow, err = decoder.DecodeFromYAML(content)
		if err != nil {
			logger.Debug("YAML parsing error: %v", err)
			return result, fmt.Errorf("error parsing YAML workflow: %w", err)
		}
	default:
		return result, fmt.Errorf("unsupported file format: %s", ext)
	}

	result.Name = workflow.Name

	var remoteWorkflow *n8n.Workflow

	if workflow.Id == nil || *workflow.Id == "" {
		result, err = CreateWorkflow(client, cmd, &workflow, filename, dryRun, result)
		if err != nil {
			return result, err
		}
		return processActivationAndTags(client, cmd, &workflow, result, dryRun)
	}

	remoteWorkflow, err = client.GetWorkflow(*workflow.Id)
	if err != nil {
		result, err = CreateWorkflowWithID(client, cmd, &workflow, filename, dryRun, result)
		if err != nil {
			return result, err
		}
		return processActivationAndTags(client, cmd, &workflow, result, dryRun)
	}

	workflowChanges := DetectWorkflowChanges(&workflow, remoteWorkflow)
	if !workflowChanges.NeedsUpdate {
		result.WorkflowID = *remoteWorkflow.Id
		status := "No content changes for"
		if !dryRun {
			status = "No changes needed for"
		}
		cmd.Printf("%s workflow '%s' (ID: %s) from %s\n", status, workflow.Name, *workflow.Id, filename)
		return processActivationAndTags(client, cmd, &workflow, result, dryRun)
	}

	result, err = UpdateWorkflow(client, cmd, &workflow, filename, dryRun, result)
	if err != nil {
		return result, err
	}

	return processActivationAndTags(client, cmd, &workflow, result, dryRun)
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

// WorkflowChange represents possible changes between local and remote workflows
type WorkflowChange struct {
	NeedsUpdate       bool
	NeedsActivation   bool
	NeedsDeactivation bool
	NeedsTagsUpdate   bool
}

// DetectWorkflowChanges compares local and remote workflows to detect what changes are needed
func DetectWorkflowChanges(local *n8n.Workflow, remote *n8n.Workflow) WorkflowChange {
	changes := WorkflowChange{}

	if remote == nil {
		if local.Active != nil && *local.Active {
			changes.NeedsActivation = true
		}
		if local.Tags != nil && len(*local.Tags) > 0 {
			changes.NeedsTagsUpdate = true
		}
		return changes
	}

	localCopy := *local
	localCopy.Id = nil
	localCopy.Active = nil
	localCopy.Tags = nil

	remoteCopy := *remote
	remoteCopy.Id = nil
	remoteCopy.Active = nil
	remoteCopy.Tags = nil

	changes.NeedsUpdate = cmd.DetectWorkflowDrift(remoteCopy, localCopy, true)

	if local.Active != nil && remote.Active != nil {
		if *local.Active && !*remote.Active {
			changes.NeedsActivation = true
		} else if !*local.Active && *remote.Active {
			changes.NeedsDeactivation = true
		}
	} else if local.Active != nil && *local.Active {
		changes.NeedsActivation = true
	}

	if local.Tags != nil && len(*local.Tags) > 0 {
		if remote.Tags == nil || len(*remote.Tags) == 0 {
			changes.NeedsTagsUpdate = true
		} else {
			if !reflect.DeepEqual(local.Tags, remote.Tags) {
				changes.NeedsTagsUpdate = true
			}
		}
	}

	return changes
}

// HandleTagUpdates updates the tags for a workflow if needed
func HandleTagUpdates(client n8n.ClientInterface, cmd *cobra.Command, workflow *n8n.Workflow, workflowID string, dryRun bool) error {
	if workflow.Tags == nil || len(*workflow.Tags) == 0 {
		return nil
	}

	var existingTags map[string]string
	var err error

	if dryRun {
		existingTags = make(map[string]string)
	} else {
		existingTags, err = getExistingTagsMap(client)
		if err != nil {
			return fmt.Errorf("error fetching existing tags: %w", err)
		}
	}

	var tagIDs n8n.TagIds
	for _, tag := range *workflow.Tags {
		if tag.Id != nil && *tag.Id != "" {
			tagIDs = append(tagIDs, struct {
				Id string `json:"id"`
			}{Id: *tag.Id})
			continue
		}

		if dryRun {
			cmd.Printf("Would create tag '%s' for workflow '%s'\n", tag.Name, workflow.Name)
			continue
		}

		if tagID, exists := existingTags[tag.Name]; exists {
			tagIDs = append(tagIDs, struct {
				Id string `json:"id"`
			}{Id: tagID})
			continue
		}

		dryRunMsg := fmt.Sprintf("Would create tag '%s' for workflow '%s'", tag.Name, workflow.Name)
		createErr := ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
			createdTag, err := client.CreateTag(tag.Name)
			if err != nil {
				return "", fmt.Errorf("error creating tag '%s': %w", tag.Name, err)
			}

			if createdTag.Id != nil {
				tagIDs = append(tagIDs, struct {
					Id string `json:"id"`
				}{Id: *createdTag.Id})
				existingTags[tag.Name] = *createdTag.Id
			}

			return fmt.Sprintf("Created tag '%s' (ID: %s)", tag.Name, *createdTag.Id), nil
		})

		if createErr != nil {
			return createErr
		}
	}

	if len(tagIDs) == 0 && !dryRun {
		return nil
	}

	dryRunMsg := fmt.Sprintf("Would update tags for workflow '%s' (ID: %s)", workflow.Name, workflowID)
	return ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
		_, err := client.UpdateWorkflowTags(workflowID, tagIDs)
		if err != nil {
			return "", fmt.Errorf("error updating workflow tags: %w", err)
		}
		return fmt.Sprintf("Updated tags for workflow '%s' (ID: %s)", workflow.Name, workflowID), nil
	})
}

// CreateWorkflow creates a new workflow without ID
func CreateWorkflow(client n8n.ClientInterface, cmd *cobra.Command, workflow *n8n.Workflow, filename string, dryRun bool, result WorkflowResult) (WorkflowResult, error) {
	dryRunMsg := fmt.Sprintf("Would create workflow '%s' from %s", workflow.Name, filename)

	err := ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
		w, err := client.CreateWorkflow(workflow)
		if err != nil {
			return "", fmt.Errorf("error creating workflow: %w", err)
		}
		result.Created = true
		result.WorkflowID = *w.Id
		return fmt.Sprintf("Created workflow '%s' (ID: %s) from %s", w.Name, *w.Id, filename), nil
	})

	return result, err
}

// CreateWorkflowWithID creates a new workflow with a specified ID
func CreateWorkflowWithID(client n8n.ClientInterface, cmd *cobra.Command, workflow *n8n.Workflow, filename string, dryRun bool, result WorkflowResult) (WorkflowResult, error) {
	dryRunMsg := fmt.Sprintf("Would create workflow '%s' with ID %s from %s (ID specified but not found on server)", workflow.Name, *workflow.Id, filename)

	err := ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
		w, err := client.CreateWorkflow(workflow)
		if err != nil {
			return "", fmt.Errorf("error creating workflow: %w", err)
		}
		result.Created = true
		result.WorkflowID = *w.Id
		return fmt.Sprintf("Created workflow '%s' (ID: %s) from %s", w.Name, *w.Id, filename), nil
	})

	return result, err
}

// UpdateWorkflow updates an existing workflow
func UpdateWorkflow(client n8n.ClientInterface, cmd *cobra.Command, workflow *n8n.Workflow, filename string, dryRun bool, result WorkflowResult) (WorkflowResult, error) {
	dryRunMsg := fmt.Sprintf("Would update workflow '%s' (ID: %s) from %s", workflow.Name, *workflow.Id, filename)

	err := ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
		w, err := client.UpdateWorkflow(*workflow.Id, workflow)
		if err != nil {
			return "", fmt.Errorf("error updating workflow: %w", err)
		}
		result.Updated = true
		result.WorkflowID = *w.Id
		return fmt.Sprintf("Updated workflow '%s' (ID: %s) from %s", w.Name, *w.Id, filename), nil
	})

	return result, err
}

// processActivationAndTags handles activation/deactivation and tag updates for a workflow
func processActivationAndTags(client n8n.ClientInterface, cmd *cobra.Command, workflow *n8n.Workflow, result WorkflowResult, dryRun bool) (WorkflowResult, error) {
	if result.WorkflowID == "" {
		return result, nil
	}

	workflowID := result.WorkflowID
	workflowName := workflow.Name

	idInfo := fmt.Sprintf("(ID: %s)", workflowID)

	var changes WorkflowChange
	if result.Created {
		if workflow.Active != nil && *workflow.Active {
			changes.NeedsActivation = true
		}
		if workflow.Tags != nil && len(*workflow.Tags) > 0 {
			changes.NeedsTagsUpdate = true
		}
	} else {
		remoteWorkflow, fetchErr := client.GetWorkflow(workflowID)
		if fetchErr != nil {
			cmd.Printf("Warning: Could not retrieve workflow details for activation/tag processing: %v\n", fetchErr)

			if workflow.Active != nil && *workflow.Active {
				changes.NeedsActivation = true
			}
			if workflow.Tags != nil && len(*workflow.Tags) > 0 {
				changes.NeedsTagsUpdate = true
			}
		} else {
			changes = DetectWorkflowChanges(workflow, remoteWorkflow)
		}
	}

	if workflow.Active != nil {
		if *workflow.Active && changes.NeedsActivation {
			dryRunMsg := fmt.Sprintf("Would activate workflow '%s' %s", workflowName, idInfo)

			activateErr := ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
				_, err := client.ActivateWorkflow(workflowID)
				if err != nil {
					return "", fmt.Errorf("error activating workflow: %w", err)
				}
				return fmt.Sprintf("Activated workflow '%s' %s", workflowName, idInfo), nil
			})

			if activateErr != nil {
				return result, activateErr
			}
		} else if !*workflow.Active && changes.NeedsDeactivation {
			dryRunMsg := fmt.Sprintf("Would deactivate workflow '%s' %s", workflowName, idInfo)

			deactivateErr := ExecuteOrDryRun(cmd, dryRun, dryRunMsg, func() (string, error) {
				_, err := client.DeactivateWorkflow(workflowID)
				if err != nil {
					return "", fmt.Errorf("error deactivating workflow: %w", err)
				}
				return fmt.Sprintf("Deactivated workflow '%s' %s", workflowName, idInfo), nil
			})

			if deactivateErr != nil {
				return result, deactivateErr
			}
		}
	}

	if changes.NeedsTagsUpdate && workflow.Tags != nil && len(*workflow.Tags) > 0 {
		if tagErr := HandleTagUpdates(client, cmd, workflow, workflowID, dryRun); tagErr != nil {
			return result, tagErr
		}
	}

	return result, nil
}

// getExistingTagsMap fetches existing tags from n8n and returns a map of tag name to tag ID
func getExistingTagsMap(client n8n.ClientInterface) (map[string]string, error) {
	tagMap := make(map[string]string)

	tagList, err := client.GetTags()
	if err != nil {
		return nil, fmt.Errorf("error fetching tags: %w", err)
	}

	if tagList != nil && tagList.Data != nil {
		for _, tag := range *tagList.Data {
			if tag.Id != nil {
				tagMap[tag.Name] = *tag.Id
			}
		}
	}

	return tagMap, nil
}
