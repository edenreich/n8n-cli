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
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ExecutionHandler handles execution history commands
type ExecutionHandler struct {
	Client n8n.ClientInterface
}

// ExecutionsCmd represents the executions command
var ExecutionsCmd = &cobra.Command{
	Use:   "executions [WORKFLOW_ID]",
	Short: "Get execution history for workflows",
	Long:  `Retrieve execution history for n8n workflows. If a workflow ID is provided, only executions for that specific workflow are returned.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, ok := viper.Get("api_key").(string)
		if !ok || apiKey == "" {
			return fmt.Errorf("API key not found in configuration")
		}

		instanceURL, ok := viper.Get("instance_url").(string)
		if !ok || instanceURL == "" {
			return fmt.Errorf("instance URL not found in configuration")
		}

		client := n8n.NewClient(instanceURL, apiKey)
		handler := ExecutionHandler{Client: client}
		return handler.Handle(cmd, args)
	},
}

func init() {
	rootcmd.GetWorkflowsCmd().AddCommand(ExecutionsCmd)

	ExecutionsCmd.Flags().BoolP("include-data", "d", false, "Include execution data in results")
	ExecutionsCmd.Flags().StringP("status", "s", "", "Filter by execution status: error, success, or waiting")
	ExecutionsCmd.Flags().IntP("limit", "l", 10, "Maximum number of executions to return")
	ExecutionsCmd.Flags().String("cursor", "", "Cursor for pagination")
	ExecutionsCmd.Flags().BoolP("json", "j", false, "Output results in JSON format")
	ExecutionsCmd.Flags().Bool("raw", false, "Output raw JSON response")
	ExecutionsCmd.Flags().BoolP("no-truncate", "n", false, "Show all nodes in the execution flow path (default: show max 5 nodes)")
	ExecutionsCmd.Flags().Int("max-nodes", 5, "Maximum number of nodes to show in the flow path (0 for all)")
}

// Handle executes the executions command
func (h ExecutionHandler) Handle(cmd *cobra.Command, args []string) error {
	includeData, _ := cmd.Flags().GetBool("include-data")
	status, _ := cmd.Flags().GetString("status")
	limit, _ := cmd.Flags().GetInt("limit")
	cursor, _ := cmd.Flags().GetString("cursor")
	outputJSON, _ := cmd.Flags().GetBool("json")
	rawJSON, _ := cmd.Flags().GetBool("raw")

	if status != "" && status != "error" && status != "success" && status != "waiting" {
		return fmt.Errorf("invalid status filter: %s. Valid values are: error, success, waiting", status)
	}

	var workflowID string
	if len(args) > 0 {
		workflowID = args[0]
	}

	executions, err := h.Client.GetExecutions(workflowID, includeData, status, limit, cursor)
	if err != nil {
		_, printErr := fmt.Fprintf(cmd.ErrOrStderr(), "Error getting executions: %v\n", err)
		if printErr != nil {
			return fmt.Errorf("failed to write error: %v (original error: %v)", printErr, err)
		}
		return err
	}

	if rawJSON {
		jsonData, err := json.MarshalIndent(executions, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal executions: %v", err)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", jsonData)
		return err
	}

	if outputJSON {
		jsonData, err := json.MarshalIndent(executions, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal executions: %v", err)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", jsonData)
		return err
	}

	if executions.Data == nil || len(*executions.Data) == 0 {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "No executions found.\n")
		return err
	}

	var header string
	if workflowID != "" {
		header = fmt.Sprintf("Execution history for workflow ID %s:\n\n", workflowID)
	} else {
		header = "Execution history for all workflows:\n\n"
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), header)
	if err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	noTruncate, _ := cmd.Flags().GetBool("no-truncate")
	maxNodes, _ := cmd.Flags().GetInt("max-nodes")

	if noTruncate {
		maxNodes = 0
	}

	if err := printExecutions(cmd.OutOrStdout(), executions, maxNodes); err != nil {
		return err
	}

	if executions.NextCursor != nil && *executions.NextCursor != "" {
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "\nMore results available. Use --cursor=%s to get the next page.\n", *executions.NextCursor)
		if err != nil {
			return fmt.Errorf("failed to write pagination info: %v", err)
		}
	}

	return nil
}

// printExecutions prints execution information in a formatted table
func printExecutions(out io.Writer, executions *n8n.ExecutionList, maxNodes int) error {
	headerFormat := "%-6s %-45s %-10s %-19s %-6s %-10s\n"
	rowFormat := "%-6s %-45.45s %-10s %-19s %-6s %-10s\n"

	_, err := fmt.Fprintf(out, headerFormat, "ID", "Flow (Node Execution Path)", "Status", "Started", "Time", "Mode")
	if err != nil {
		return fmt.Errorf("failed to write table header: %v", err)
	}

	_, err = fmt.Fprintf(out, "%s\n", strings.Repeat("-", 105))
	if err != nil {
		return fmt.Errorf("failed to write table separator: %v", err)
	}

	for _, execution := range *executions.Data {
		id := "N/A"
		if execution.Id != nil {
			id = strconv.FormatFloat(float64(*execution.Id), 'f', 0, 32)
		}

		flowPath := "N/A"

		if execution.Data != nil {
			if resultData, ok := (*execution.Data)["resultData"].(map[string]interface{}); ok {
				var lastNode string
				if lastNodeStr, ok := resultData["lastNodeExecuted"].(string); ok {
					lastNode = lastNodeStr
				}

				if runData, ok := resultData["runData"].(map[string]interface{}); ok {
					type NodeExecution struct {
						Name      string
						StartTime int64
					}

					var nodeExecutions []NodeExecution
					for nodeName, nodeData := range runData {
						if dataArray, ok := nodeData.([]interface{}); ok && len(dataArray) > 0 {
							if execution, ok := dataArray[0].(map[string]interface{}); ok {
								if startTime, ok := execution["startTime"].(float64); ok {
									nodeExecutions = append(nodeExecutions, NodeExecution{
										Name:      nodeName,
										StartTime: int64(startTime),
									})
								}
							}
						}
					}

					sort.Slice(nodeExecutions, func(i, j int) bool {
						return nodeExecutions[i].StartTime < nodeExecutions[j].StartTime
					})

					var flowNodes []string
					for _, node := range nodeExecutions {
						if node.Name == lastNode {
							flowNodes = append(flowNodes, node.Name+"*")
						} else {
							flowNodes = append(flowNodes, node.Name)
						}
					}

					if len(flowNodes) > 0 {
						if maxNodes <= 0 || len(flowNodes) <= maxNodes {
							flowPath = strings.Join(flowNodes, " → ")
						} else if maxNodes == 1 {
							if flowNodes[0] == lastNode+"*" {
								flowPath = flowNodes[0]
							} else {
								flowPath = flowNodes[0] + " → ..."
							}
						} else {
							firstNodes := maxNodes - 2
							if firstNodes < 1 {
								firstNodes = 1
							}
							flowPath = strings.Join(flowNodes[:firstNodes], " → ") +
								" → ... → " + flowNodes[len(flowNodes)-1]
						}
					}
				}
			}
		}

		if execution.WorkflowId != nil {
			workflowId := strconv.FormatFloat(float64(*execution.WorkflowId), 'f', 0, 32)
			if flowPath == "N/A" {
				flowPath = fmt.Sprintf("ID: %s", workflowId)
			}
		}

		status := "N/A"
		if execution.Finished != nil {
			if *execution.Finished {
				status = "Success"
			} else {
				status = "Running"
			}
		}

		if execution.WaitTill != nil && execution.Finished != nil && !*execution.Finished {
			status = "Waiting"
		}

		startedAt := "N/A"
		duration := "N/A"
		if execution.StartedAt != nil {
			startedAt = execution.StartedAt.Format(time.RFC3339)[:19]

			if execution.StoppedAt != nil {
				durationMs := execution.StoppedAt.Sub(*execution.StartedAt).Milliseconds()
				if durationMs < 1000 {
					duration = fmt.Sprintf("%dms", durationMs)
				} else {
					duration = fmt.Sprintf("%.1fs", float64(durationMs)/1000)
				}
			}
		}

		mode := "N/A"
		if execution.Mode != nil {
			mode = string(*execution.Mode)
		}

		_, err := fmt.Fprintf(out, rowFormat, id, flowPath, status, startedAt, duration, mode)
		if err != nil {
			return fmt.Errorf("failed to write execution row: %v", err)
		}
	}

	return nil
}
