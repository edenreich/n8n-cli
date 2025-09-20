// Package n8n provides primitives to interact with the openapi HTTP API.
package n8n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edenreich/n8n-cli/logger"
	"gopkg.in/yaml.v3"
)

// CleanWorkflow creates a clean copy of a workflow with null and empty fields removed.
// It preserves essential fields like typeVersion for proper node compatibility.
func CleanWorkflow(workflow Workflow) Workflow {
	cleanedWorkflow := workflow

	cleanedWorkflow.CreatedAt = nil
	cleanedWorkflow.UpdatedAt = nil
	cleanedWorkflow.Shared = nil

	if cleanedWorkflow.Tags != nil && len(*cleanedWorkflow.Tags) > 0 {
		cleanTags := make([]Tag, len(*cleanedWorkflow.Tags))
		for i, tag := range *cleanedWorkflow.Tags {
			cleanTags[i] = Tag{
				Id:   tag.Id,
				Name: tag.Name,
			}
		}
		cleanedWorkflow.Tags = &cleanTags
	}

	if cleanedWorkflow.Connections == nil {
		cleanedWorkflow.Connections = make(map[string]interface{})
	}

	return cleanedWorkflow
}

// WorkflowEncoder handles various encoding formats for n8n workflows
type WorkflowEncoder struct {
	Clean bool
}

// NewWorkflowEncoder creates a new encoder with the specified options
func NewWorkflowEncoder(clean bool) *WorkflowEncoder {
	return &WorkflowEncoder{
		Clean: clean,
	}
}

// EncodeToJSON encodes a workflow to a JSON byte array
func (e *WorkflowEncoder) EncodeToJSON(workflow Workflow) ([]byte, error) {
	var workflowToEncode Workflow

	if e.Clean {
		workflowToEncode = CleanWorkflow(workflow)
	} else {
		workflowToEncode = workflow
	}

	return json.MarshalIndent(workflowToEncode, "", "  ")
}

// EncodeToYAML encodes a workflow to a YAML byte array with proper formatting
func (e *WorkflowEncoder) EncodeToYAML(workflow Workflow) ([]byte, error) {
	var workflowToEncode Workflow

	if e.Clean {
		workflowToEncode = CleanWorkflow(workflow)
	} else {
		workflowToEncode = workflow
	}

	jsonData, err := json.Marshal(workflowToEncode)
	if err != nil {
		return nil, fmt.Errorf("failed to encode workflow to JSON before YAML conversion: %w", err)
	}

	var workflowMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &workflowMap); err != nil {
		return nil, fmt.Errorf("failed to convert JSON to map for YAML encoding: %w", err)
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(workflowMap); err != nil {
		return nil, fmt.Errorf("failed to encode workflow to YAML: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize YAML encoding: %w", err)
	}

	return append([]byte("---\n"), buf.Bytes()...), nil
}

type WorkflowDecoder struct{}

// NewWorkflowDecoder creates a new decoder
func NewWorkflowDecoder() *WorkflowDecoder {
	return &WorkflowDecoder{}
}

// DecodeFromJSON decodes a workflow from a JSON byte array
func (d *WorkflowDecoder) DecodeFromJSON(data []byte) (Workflow, error) {
	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return Workflow{}, fmt.Errorf("failed to decode workflow from JSON: %w", err)
	}
	return workflow, nil
}

// DecodeFromYAML decodes a workflow from a YAML byte array
func (d *WorkflowDecoder) DecodeFromYAML(data []byte) (Workflow, error) {
	logger.Debug("YAML INPUT:\n%s", string(data))

	var workflowMap map[string]interface{}
	if err := yaml.Unmarshal(data, &workflowMap); err != nil {
		return Workflow{}, fmt.Errorf("failed to decode workflow from YAML: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(workflowMap, "", "  ")
	if err == nil {
		logger.Debug("YAML TO JSON INTERMEDIATE MAP:\n%s", string(jsonBytes))
	}

	jsonData, err := json.Marshal(workflowMap)
	if err != nil {
		return Workflow{}, fmt.Errorf("failed to convert YAML map to JSON: %w", err)
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonData, "", "  "); err == nil {
		logger.Debug("YAML TO JSON FINAL SERIALIZED OUTPUT:\n%s", prettyJSON.String())
	} else {
		logger.Debug("YAML TO JSON FINAL OUTPUT (could not prettify): %s", string(jsonData))
	}

	logger.Debug("YAML TO JSON (AFTER MARSHAL): %s", string(jsonData))

	var workflow Workflow
	if err := json.Unmarshal(jsonData, &workflow); err != nil {
		logger.Debug("ERROR UNMARSHALING JSON TO WORKFLOW: %v", err)

		var anyMap map[string]interface{}
		if mapErr := json.Unmarshal(jsonData, &anyMap); mapErr == nil {
			prettyJSON, _ := json.MarshalIndent(anyMap, "", "  ")
			logger.Debug("JSON STRUCTURE THAT FAILED TO UNMARSHAL:\n%s", string(prettyJSON))
		}
		return Workflow{}, fmt.Errorf("failed to convert JSON to workflow: %w", err)
	}

	return workflow, nil
}

// DecodeFromBytes attempts to decode a workflow from bytes, with smart format detection
func (d *WorkflowDecoder) DecodeFromBytes(data []byte) (Workflow, error) {
	trimmed := string(data)
	maxChars := 50
	if len(trimmed) > maxChars {
		trimmed = trimmed[:maxChars]
	}

	if strings.HasPrefix(strings.TrimSpace(trimmed), "{") || strings.HasPrefix(strings.TrimSpace(trimmed), "[") {
		return d.DecodeFromJSON(data)
	}

	if strings.HasPrefix(strings.TrimSpace(trimmed), "---") ||
		(!strings.Contains(trimmed, "{") && !strings.Contains(trimmed, "[")) {
		return d.DecodeFromYAML(data)
	}

	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err == nil {
		return workflow, nil
	}

	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return Workflow{}, fmt.Errorf("failed to decode workflow from JSON or YAML: %w", err)
	}

	return workflow, nil
}
