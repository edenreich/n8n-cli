package n8n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client is a simple client for interacting with n8n API
type Client struct {
	baseURL  string
	apiToken string
	client   *http.Client
}

// NewClient creates a new n8n client
func NewClient(baseURL, apiToken string) *Client {
	return &Client{
		baseURL:  baseURL + "/api/v1",
		apiToken: apiToken,
		client:   &http.Client{},
	}
}

// GetWorkflows fetches workflows from the n8n API
func (c *Client) GetWorkflows() (*WorkflowList, error) {
	url := fmt.Sprintf("%s/workflows", c.baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result WorkflowList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ActivateWorkflow activates a workflow by ID
func (c *Client) ActivateWorkflow(id string) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows/%s/activate", c.baseURL, id)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result Workflow
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeactivateWorkflow deactivates a workflow by ID
func (c *Client) DeactivateWorkflow(id string) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows/%s/deactivate", c.baseURL, id)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result Workflow
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateWorkflow creates a new workflow
func (c *Client) CreateWorkflow(workflow *Workflow) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows", c.baseURL)

	body, err := json.Marshal(workflow)
	if err != nil {
		return nil, fmt.Errorf("error marshaling workflow: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var w Workflow
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, err
	}

	return &w, nil
}

// UpdateWorkflow updates an existing workflow by its ID
func (c *Client) UpdateWorkflow(id string, workflow *Workflow) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.baseURL, id)

	workflowCopy := *workflow
	workflowCopy.Id = nil
	workflowCopy.Active = nil
	workflowCopy.CreatedAt = nil
	workflowCopy.UpdatedAt = nil
	workflowCopy.Tags = nil

	body, err := json.Marshal(workflowCopy)
	if err != nil {
		return nil, fmt.Errorf("error marshaling workflow: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var w Workflow
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, err
	}

	return &w, nil
}

// DeleteWorkflow deletes a workflow by ID
func (c *Client) DeleteWorkflow(id string) error {
	url := fmt.Sprintf("%s/workflows/%s", c.baseURL, id)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}
