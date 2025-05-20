package n8n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Client is a simple client for interacting with n8n API
type Client struct {
	baseURL  string
	apiToken string
	client   *http.Client
	logger   *zap.SugaredLogger
}

// NewClient creates a new n8n client
func NewClient(baseURL, apiToken string) *Client {
	var logger *zap.SugaredLogger

	if os.Getenv("DEBUG") == "1" || os.Getenv("DEBUG") == "true" {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		zapLogger, err := cfg.Build()
		if err != nil {
			zapLogger, _ = zap.NewProduction()
		}

		logger = zapLogger.Sugar().Named("n8n-api")
	} else {
		zapLogger, _ := zap.NewProduction()
		logger = zapLogger.Sugar().Named("n8n-api")
	}

	return &Client{
		baseURL:  baseURL + "/api/v1",
		apiToken: apiToken,
		client:   &http.Client{},
		logger:   logger,
	}
}

// logDebug logs a debug message
func (c *Client) logDebug(format string, args ...interface{}) {
	c.logger.Debugf(format, args...)
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
			c.logger.Warnf("Error closing response body: %v", err)
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
			c.logger.Warnf("Error closing response body: %v", err)
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
			c.logger.Warnf("Error closing response body: %v", err)
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

	c.logDebug("CREATE WORKFLOW REQUEST: %s", string(body))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		c.logDebug("CREATE WORKFLOW FORMATTED JSON:\n%s", prettyJSON.String())
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	c.logDebug("CREATE/UPDATE WORKFLOW RESPONSE (Status: %d): %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, respBody)
	}

	var w Workflow
	if err := json.NewDecoder(bytes.NewBuffer(respBody)).Decode(&w); err != nil {
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

	c.logDebug("UPDATE WORKFLOW REQUEST (ID: %s): %s", id, string(body))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		c.logDebug("UPDATE WORKFLOW FORMATTED JSON (ID: %s):\n%s", id, prettyJSON.String())
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
			c.logger.Warnf("Error closing response body: %v", err)
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

// GetWorkflow fetches a single workflow by its ID
func (c *Client) GetWorkflow(id string) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.baseURL, id)

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
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var workflow Workflow
	if err := json.NewDecoder(resp.Body).Decode(&workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
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
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}

// GetWorkflowTags fetches the tags of a workflow by its ID
func (c *Client) GetWorkflowTags(id string) (WorkflowTags, error) {
	url := fmt.Sprintf("%s/workflows/%s/tags", c.baseURL, id)

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
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var tags WorkflowTags
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// UpdateWorkflowTags updates the tags of a workflow by its ID
func (c *Client) UpdateWorkflowTags(id string, tagIds TagIds) (WorkflowTags, error) {
	url := fmt.Sprintf("%s/workflows/%s/tags", c.baseURL, id)

	jsonBody, err := json.Marshal(tagIds)
	if err != nil {
		return nil, err
	}

	c.logDebug("UPDATE WORKFLOW TAGS REQUEST (ID: %s): %s", id, string(jsonBody))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonBody, "", "  "); err == nil {
		c.logDebug("UPDATE WORKFLOW TAGS FORMATTED JSON (ID: %s):\n%s", id, prettyJSON.String())
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
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
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var tags WorkflowTags
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// CreateTag creates a new tag in n8n
func (c *Client) CreateTag(tagName string) (*Tag, error) {
	url := fmt.Sprintf("%s/tags", c.baseURL)

	tag := Tag{
		Name: tagName,
	}

	body, err := json.Marshal(tag)
	if err != nil {
		return nil, fmt.Errorf("error marshaling tag: %w", err)
	}

	c.logDebug("CREATE TAG REQUEST: %s", string(body))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		c.logDebug("CREATE TAG FORMATTED JSON:\n%s", prettyJSON.String())
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	c.logDebug("CREATE TAG RESPONSE (Status: %d): %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, respBody)
	}

	var createdTag Tag
	if err := json.NewDecoder(bytes.NewBuffer(respBody)).Decode(&createdTag); err != nil {
		return nil, err
	}

	return &createdTag, nil
}

// GetTags fetches all tags from n8n
func (c *Client) GetTags() (*TagList, error) {
	url := fmt.Sprintf("%s/tags", c.baseURL)

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
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var tags TagList
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	return &tags, nil
}
