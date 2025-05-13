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
	"github.com/edenreich/n8n-cli/cmd"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import JSON workflows into n8n instance",
	Long:  `Import command imports workflows from n8n instance.`,
	RunE:  importWorkflows,
}

func init() {
	cmd.GetWorkflowsCmd().AddCommand(ImportCmd)
}

// importWorkflows imports workflows from n8n instance
func importWorkflows(cmd *cobra.Command, args []string) error {
	cmd.Println("Importing workflows...")
	// TODO - implement import command here

	return nil
}
