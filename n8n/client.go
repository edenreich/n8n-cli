package n8n

import (
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
		baseURL:  baseURL,
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
