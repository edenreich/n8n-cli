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
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"text/tabwriter"

	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
)

var workflowsCmd = &cobra.Command{
	Use:   "workflows",
	Short: "Manage n8n workflows",
	Long: `Manage n8n workflows.
	
This command provides subcommands to manage your n8n workflows, including 
activating, deactivating, and listing workflows.`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workflows",
	Long:  `List all workflows from your n8n instance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get config: %v", err)
		}

		workflows, err := getServerWorkflows(cfg.APIBaseURL, cfg.APIToken)
		if err != nil {
			return fmt.Errorf("failed to get workflows: %v", err)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		defer func() {
			if err := w.Flush(); err != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error flushing output: %v\n", err)
			}
		}()

		if _, err := fmt.Fprintln(w, "ID\tNAME\tACTIVE\t"); err != nil {
			return fmt.Errorf("error writing header: %v", err)
		}

		for _, workflow := range workflows {
			id := *workflow.Id
			name := workflow.Name
			active := *workflow.Active
			if _, err := fmt.Fprintf(w, "%v\t%v\t%v\t\n", id, name, active); err != nil {
				return fmt.Errorf("error writing workflow data: %v", err)
			}
		}

		return nil
	},
}

var activateCmd = &cobra.Command{
	Use:   "activate [workflow_id]",
	Short: "Activate a workflow",
	Long: `Activate a workflow in n8n.

This command activates a workflow in your n8n instance by its ID or name.
When a workflow is activated, it will run according to its trigger settings.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowIdentifier := args[0]
		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get config: %v", err)
		}

		if err := setWorkflowActiveState(workflowIdentifier, true, cfg.APIBaseURL, cfg.APIToken); err != nil {
			return fmt.Errorf("failed to activate workflow: %v", err)
		}

		cmd.Printf("Workflow '%s' has been activated\n", workflowIdentifier)
		return nil
	},
}

var deactivateCmd = &cobra.Command{
	Use:   "deactivate [workflow_id]",
	Short: "Deactivate a workflow",
	Long: `Deactivate a workflow in n8n.

This command deactivates a workflow in your n8n instance by its ID or name.
When a workflow is deactivated, it will not run automatically even if it has triggers.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workflowIdentifier := args[0]
		cfg, err := config.GetConfig()
		if err != nil {
			return fmt.Errorf("failed to get config: %v", err)
		}

		if err := setWorkflowActiveState(workflowIdentifier, false, cfg.APIBaseURL, cfg.APIToken); err != nil {
			return fmt.Errorf("failed to deactivate workflow: %v", err)
		}

		cmd.Printf("Workflow '%s' has been deactivated\n", workflowIdentifier)
		return nil
	},
}

func setWorkflowActiveState(workflowIdentifier string, active bool, apiBaseURL, apiToken string) error {
	workflow, err := findWorkflow(workflowIdentifier, apiBaseURL, apiToken)
	if err != nil {
		return err
	}

	var endpoint string
	if active {
		endpoint = "activate"
	} else {
		endpoint = "deactivate"
	}

	url := fmt.Sprintf("%s/workflows/%s/%s", apiBaseURL, *workflow.Id, endpoint)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// searchWorkflowsByName searches for workflows by name
func searchWorkflowsByName(name string, apiBaseURL, apiToken string) ([]*n8n.Workflow, error) {
	endpoint := fmt.Sprintf("%s/workflows?name=%s", apiBaseURL, url.QueryEscape(name))

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-N8N-API-KEY", apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response struct {
		Data []n8n.Workflow `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode workflows: %w", err)
	}

	workflows := make([]*n8n.Workflow, len(response.Data))
	for i := range response.Data {
		workflows[i] = &response.Data[i]
	}

	return workflows, nil
}

// getWorkflowByID retrieves a workflow by its ID
func getWorkflowByID(id string, apiBaseURL, apiToken string) (*n8n.Workflow, error) {
	endpoint := fmt.Sprintf("%s/workflows/%s", apiBaseURL, id)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-N8N-API-KEY", apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var workflow n8n.Workflow
	if err := json.NewDecoder(resp.Body).Decode(&workflow); err != nil {
		return nil, fmt.Errorf("failed to decode workflow: %w", err)
	}

	return &workflow, nil
}

// findWorkflow searches for a workflow by ID or name
func findWorkflow(identifier string, apiBaseURL, apiToken string) (*n8n.Workflow, error) {
	workflow, err := getWorkflowByID(identifier, apiBaseURL, apiToken)
	if err == nil {
		return workflow, nil
	}

	workflows, err := searchWorkflowsByName(identifier, apiBaseURL, apiToken)
	if err != nil {
		return nil, err
	}

	if len(workflows) == 0 {
		return nil, fmt.Errorf("no workflow found with identifier '%s'", identifier)
	}

	for _, wf := range workflows {
		if wf.Name == identifier {
			return wf, nil
		}
	}

	if len(workflows) == 1 {
		return workflows[0], nil
	}

	return nil, fmt.Errorf("multiple workflows found matching '%s'", identifier)
}

// getServerWorkflows retrieves all workflows from the n8n server
func getServerWorkflows(apiBaseURL, apiToken string) ([]n8n.Workflow, error) {
	endpoint := fmt.Sprintf("%s/workflows", apiBaseURL)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-N8N-API-KEY", apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Data []n8n.Workflow `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return response.Data, nil
}

func init() {
	rootCmd.AddCommand(workflowsCmd)
	workflowsCmd.AddCommand(listCmd)
	workflowsCmd.AddCommand(activateCmd)
	workflowsCmd.AddCommand(deactivateCmd)
}
