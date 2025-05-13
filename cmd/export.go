package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
)

// ImportWorkflowByIDWithConfig imports a specific workflow by ID using the provided config
func ImportWorkflowByIDWithConfig(cfg config.ConfigInterface, outputDir string, workflowID string, dryRun bool, verbose bool) error {
	url := fmt.Sprintf("%s/workflows/%s", cfg.GetAPIBaseURL(), workflowID)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-N8N-API-KEY", cfg.GetAPIToken())
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var workflow map[string]interface{}
	if err := json.Unmarshal(body, &workflow); err != nil {
		return fmt.Errorf("failed to parse workflow response: %w", err)
	}

	var workflowData map[string]interface{}
	if data, ok := workflow["data"].(map[string]interface{}); ok {
		workflowData = data
	} else {
		workflowData = workflow
	}

	workflowName, _ := workflowData["name"].(string)
	if workflowName == "" {
		workflowName = fmt.Sprintf("workflow-%s", workflowID)
	}

	filename := SanitizeFilename(workflowName) + ".json"
	filePath := filepath.Join(outputDir, filename)

	if verbose {
		fmt.Printf("Importing workflow: %s (ID: %s) to %s\n", workflowName, workflowID, filePath)
	} else {
		fmt.Printf("Importing workflow: %s\n", workflowName)
	}

	if !dryRun {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", outputDir, err)
		}

		prettyJSON, err := json.MarshalIndent(workflowData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format workflow JSON: %w", err)
		}

		prettyJSON = append(prettyJSON, '\n')

		err = os.WriteFile(filePath, prettyJSON, 0644)
		if err != nil {
			return fmt.Errorf("failed to write workflow file %s: %w", filePath, err)
		}
	}

	return nil
}

// GetServerWorkflows fetches all workflows from the n8n server
func GetServerWorkflows(baseURL string, apiToken string) ([]*n8n.Workflow, error) {
	url := fmt.Sprintf("%s/workflows", baseURL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-N8N-API-KEY", apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Data []*n8n.Workflow `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse workflow response: %w", err)
	}

	return response.Data, nil
}

// SetWorkflowActiveState activates or deactivates a workflow
func SetWorkflowActiveState(baseURL string, apiToken string, workflowID string, activate bool) error {
	action := "activate"
	if !activate {
		action = "deactivate"
	}

	url := fmt.Sprintf("%s/workflows/%s/%s", baseURL, workflowID, action)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-N8N-API-KEY", apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}
