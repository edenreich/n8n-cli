package unit

import (
	"strings"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/stretchr/testify/assert"
)

func TestCleanWorkflow(t *testing.T) {
	createdAt := timePtr("2023-01-01T00:00:00Z")
	updatedAt := timePtr("2023-01-02T00:00:00Z")

	typeVersion := float32(2.0)
	node := n8n.Node{
		Name:        stringPtr("Test Node"),
		Type:        stringPtr("n8n-nodes-base.ai.Agent"),
		TypeVersion: &typeVersion,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	workflow := n8n.Workflow{
		Name:        "Test Workflow",
		Nodes:       []n8n.Node{node},
		Connections: nil,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	cleanedWorkflow := n8n.CleanWorkflow(workflow)

	assert.Len(t, cleanedWorkflow.Nodes, 1, "Should have one node")

	assert.Nil(t, cleanedWorkflow.CreatedAt, "CreatedAt should be nil")
	assert.Nil(t, cleanedWorkflow.UpdatedAt, "UpdatedAt should be nil")

	assert.NotNil(t, cleanedWorkflow.Connections, "Connections should be initialized")

	assert.NotNil(t, cleanedWorkflow.Nodes[0].TypeVersion, "typeVersion should be preserved")
	assert.Equal(t, float32(2.0), *cleanedWorkflow.Nodes[0].TypeVersion, "typeVersion value should be 2.0")
}

func TestWorkflowEncoder(t *testing.T) {
	typeVersion := float32(2.0)
	node := n8n.Node{
		Name:        stringPtr("Test Node"),
		Type:        stringPtr("n8n-nodes-base.ai.Agent"),
		TypeVersion: &typeVersion,
	}

	workflow := n8n.Workflow{
		Name:        "Test Workflow",
		Nodes:       []n8n.Node{node},
		Connections: make(map[string]interface{}),
	}

	t.Run("JSON Encoding", func(t *testing.T) {
		encoder := n8n.NewWorkflowEncoder(true)
		jsonData, err := encoder.EncodeToJSON(workflow)

		assert.NoError(t, err, "JSON encoding should not error")
		assert.Contains(t, string(jsonData), "Test Workflow", "JSON should contain workflow name")
		assert.Contains(t, string(jsonData), "Test Node", "JSON should contain node name")
		assert.Contains(t, string(jsonData), "typeVersion", "JSON should preserve typeVersion")
	})

	t.Run("YAML Encoding", func(t *testing.T) {
		encoder := n8n.NewWorkflowEncoder(true)
		yamlData, err := encoder.EncodeToYAML(workflow)

		assert.NoError(t, err, "YAML encoding should not error")
		assert.Contains(t, string(yamlData), "Test Workflow", "YAML should contain workflow name")
		assert.Contains(t, string(yamlData), "Test Node", "YAML should contain node name")
		assert.Contains(t, string(yamlData), "typeVersion", "YAML should preserve typeVersion")
	})

	t.Run("YAML Encoding Should Have Single Separator", func(t *testing.T) {
		encoder := n8n.NewWorkflowEncoder(true)
		yamlData, err := encoder.EncodeToYAML(workflow)

		assert.NoError(t, err, "YAML encoding should not error")

		count := strings.Count(string(yamlData), "---")
		assert.Equal(t, 1, count, "YAML should contain exactly one separator")

		assert.True(t, strings.HasPrefix(string(yamlData), "---\n"), "YAML should start with a separator")
	})
}

func TestWorkflowDecoder(t *testing.T) {
	jsonData := []byte(`{
		"name": "Test JSON Workflow",
		"nodes": [
			{
				"name": "Test Node",
				"type": "n8n-nodes-base.ai.Agent",
				"typeVersion": 2.0
			}
		],
		"connections": {}
	}`)

	yamlData := []byte(`
name: Test YAML Workflow
nodes:
- name: Test Node
  type: n8n-nodes-base.ai.Agent
  typeVersion: 2.0
connections: {}
`)

	decoder := n8n.NewWorkflowDecoder()

	t.Run("JSON Decoding", func(t *testing.T) {
		workflow, err := decoder.DecodeFromJSON(jsonData)

		assert.NoError(t, err, "JSON decoding should not error")
		assert.Equal(t, "Test JSON Workflow", workflow.Name, "Workflow name should match")
		assert.Equal(t, "Test Node", *workflow.Nodes[0].Name, "Node name should match")
		assert.Equal(t, float32(2.0), *workflow.Nodes[0].TypeVersion, "typeVersion should match")
	})

	t.Run("YAML Decoding", func(t *testing.T) {
		workflow, err := decoder.DecodeFromYAML(yamlData)

		assert.NoError(t, err, "YAML decoding should not error")
		assert.Equal(t, "Test YAML Workflow", workflow.Name, "Workflow name should match")
		assert.Equal(t, "Test Node", *workflow.Nodes[0].Name, "Node name should match")
		assert.Equal(t, float32(2.0), *workflow.Nodes[0].TypeVersion, "typeVersion should match")
	})

	t.Run("Auto Detection", func(t *testing.T) {
		workflow1, err := decoder.DecodeFromBytes(jsonData)
		assert.NoError(t, err, "JSON auto-detection should not error")
		assert.Equal(t, "Test JSON Workflow", workflow1.Name, "Workflow name should match for JSON")

		workflow2, err := decoder.DecodeFromBytes(yamlData)
		assert.NoError(t, err, "YAML auto-detection should not error")
		assert.Equal(t, "Test YAML Workflow", workflow2.Name, "Workflow name should match for YAML")
	})
}
