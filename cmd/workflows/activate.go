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

// ActivateCommand represents the command to activate a workflow
type ActivateCommand struct {
	Out    io.Writer
	ErrOut io.Writer
	Client *n8n.Client
}

// ActivateCmd represents the activate command
var ActivateCmd = &cobra.Command{
	Use:   "activate WORKFLOW_ID",
	Short: "Activate a workflow by ID",
	Long:  `Activate a workflow in n8n by its ID, making it ready to be triggered by events.`,
	Args:  cobra.ExactArgs(1),
	RunE:  activateWorkflow,
}

func init() {
	rootcmd.GetWorkflowsCmd().AddCommand(ActivateCmd)
}

// activateWorkflow is the handler for the activate command
func activateWorkflow(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("this command requires a workflow ID")
	}

	apiKey := viper.Get("api_key").(string)
	instanceURL := viper.Get("instance_url").(string)

	client := n8n.NewClient(instanceURL, apiKey)

	workflowID := args[0]
	workflow, err := client.ActivateWorkflow(workflowID)
	if err != nil {
		_, printErr := fmt.Fprintf(cmd.ErrOrStderr(), "Error activating workflow: %v\n", err)
		if printErr != nil {
			return fmt.Errorf("failed to write error: %v (original error: %v)", printErr, err)
		}
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Workflow with ID %s has been activated successfully\n", workflowID)
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
