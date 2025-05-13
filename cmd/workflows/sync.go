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

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync JSON workflows to n8n instance",
	Long: `Sync command takes JSON workflow files from a specified directory 
and synchronizes them with your n8n instance. For example:

n8n-cli workflows sync --directory workflows/

This will take all JSON workflow files from workflows directory and 
upload them to the n8n instance specified in your environment configuration.

The command will:
1. Process all JSON files in the specified directory
2. For each workflow file:
   - If the workflow ID exists and is found on the n8n instance, it will update it
   - If the workflow doesn't exist or has no ID, it will create a new workflow
   - If the original workflow has "active": true, it will activate the workflow`,
	RunE: syncWorkflows,
}

func init() {
	cmd.GetWorkflowsCmd().AddCommand(SyncCmd)

	SyncCmd.Flags().StringP("directory", "d", "", "Directory containing workflow JSON files (required)")
	SyncCmd.Flags().BoolP("activate-all", "a", false, "Activate all workflows after synchronization")
	SyncCmd.Flags().Bool("dry-run", false, "Show what would be uploaded without making changes")

	// nolint:errcheck
	SyncCmd.MarkFlagRequired("directory")
}

// syncWorkflows syncs workflow files from a directory to n8n
func syncWorkflows(cmd *cobra.Command, args []string) error {
	cmd.Println("Syncing workflows...")
	directory, _ := cmd.Flags().GetString("directory")
	activateAll, _ := cmd.Flags().GetBool("activate-all")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	files, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(directory, file.Name())
			if err = processWorkflowFile(filePath, activateAll, dryRun); err != nil {
				return fmt.Errorf("error processing workflow file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

func processWorkflowFile(filePath string, activateAll, dryRun bool) error {
	// TODO: Implement the sync logic here
	return nil
}
