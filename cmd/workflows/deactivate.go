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
	"io"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DeactivateCommand represents the command to deactivate a workflow
type DeactivateCommand struct {
	Out    io.Writer
	ErrOut io.Writer
	Client *n8n.Client
}

// DeactivateCmd represents the deactivate command
var DeactivateCmd = &cobra.Command{
	Use:   "deactivate WORKFLOW_ID",
	Short: "Deactivate a workflow by ID",
	Long:  `Deactivate a workflow in n8n by its ID, stopping it from being triggered by events.`,
	Args:  cobra.ExactArgs(1),
	RunE:  deactivateWorkflow,
}

func init() {
	rootcmd.GetWorkflowsCmd().AddCommand(DeactivateCmd)
}

// deactivateWorkflow is the handler for the deactivate command
func deactivateWorkflow(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("this command requires a workflow ID")
	}

	apiKey := viper.Get("api_key").(string)
	instanceURL := viper.Get("instance_url").(string)

	client := n8n.NewClient(instanceURL, apiKey)

	workflowID := args[0]
	workflow, err := client.DeactivateWorkflow(workflowID)
	if err != nil {
		_, printErr := fmt.Fprintf(cmd.ErrOrStderr(), "Error deactivating workflow: %v\n", err)
		if printErr != nil {
			return fmt.Errorf("failed to write error: %v (original error: %v)", printErr, err)
		}
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Workflow with ID %s has been deactivated successfully\n", workflowID)
	if err != nil {
		return fmt.Errorf("failed to write output: %v", err)
	}

	if workflow.Name != "" {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", workflow.Name)
		if err != nil {
			return fmt.Errorf("failed to write workflow name: %v", err)
		}
	}

	return nil
}
