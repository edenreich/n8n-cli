package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/stretchr/testify/assert"
)

func TestCleanWorkflow_TagsCleaning(t *testing.T) {
	testTime := time.Now()

	workflow := n8n.Workflow{
		Name: "Test Workflow With Tags",
		Tags: &n8n.WorkflowTags{
			{
				Id:        stringPtr("1"),
				Name:      "tag1",
				CreatedAt: &testTime,
				UpdatedAt: &testTime,
			},
			{
				Id:        stringPtr("2"),
				Name:      "tag2",
				CreatedAt: &testTime,
				UpdatedAt: &testTime,
			},
		},
	}

	cleanedWorkflow := n8n.CleanWorkflow(workflow)

	assert.NotNil(t, cleanedWorkflow.Tags, "Tags should be preserved")
	assert.Equal(t, 2, len(*cleanedWorkflow.Tags), "Number of tags should remain the same")

	assert.Equal(t, "1", *(*cleanedWorkflow.Tags)[0].Id, "Tag ID should be preserved")
	assert.Equal(t, "tag1", (*cleanedWorkflow.Tags)[0].Name, "Tag name should be preserved")
	assert.Nil(t, (*cleanedWorkflow.Tags)[0].CreatedAt, "Tag CreatedAt should be nil")
	assert.Nil(t, (*cleanedWorkflow.Tags)[0].UpdatedAt, "Tag UpdatedAt should be nil")

	assert.Equal(t, "2", *(*cleanedWorkflow.Tags)[1].Id, "Tag ID should be preserved")
	assert.Equal(t, "tag2", (*cleanedWorkflow.Tags)[1].Name, "Tag name should be preserved")
	assert.Nil(t, (*cleanedWorkflow.Tags)[1].CreatedAt, "Tag CreatedAt should be nil")
	assert.Nil(t, (*cleanedWorkflow.Tags)[1].UpdatedAt, "Tag UpdatedAt should be nil")
}

func TestWorkflowEncoder_WithTags(t *testing.T) {
	testTime := time.Now()

	id1 := "1"
	id2 := "2"

	workflow := n8n.Workflow{
		Name: "Test Workflow With Tags",
		Tags: &n8n.WorkflowTags{
			{
				Id:        &id1,
				Name:      "tag1",
				CreatedAt: &testTime,
				UpdatedAt: &testTime,
			},
			{
				Id:        &id2,
				Name:      "tag2",
				CreatedAt: &testTime,
				UpdatedAt: &testTime,
			},
		},
	}

	t.Run("With minimal=true", func(t *testing.T) {
		encoder := n8n.NewWorkflowEncoder(true)
		jsonData, err := encoder.EncodeToJSON(workflow)
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(jsonData, &result)
		assert.NoError(t, err)

		assert.Contains(t, result, "tags")
		tags := result["tags"].([]interface{})
		assert.Len(t, tags, 2)

		tag1 := tags[0].(map[string]interface{})
		assert.Contains(t, tag1, "id")
		assert.Contains(t, tag1, "name")
		assert.NotContains(t, tag1, "createdAt")
		assert.NotContains(t, tag1, "updatedAt")
	})

	t.Run("With minimal=false", func(t *testing.T) {
		encoder := n8n.NewWorkflowEncoder(false)
		jsonData, err := encoder.EncodeToJSON(workflow)
		assert.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(jsonData, &result)
		assert.NoError(t, err)

		assert.Contains(t, result, "tags")
		tags := result["tags"].([]interface{})
		assert.Len(t, tags, 2)

		tag1 := tags[0].(map[string]interface{})
		assert.Contains(t, tag1, "id")
		assert.Contains(t, tag1, "name")
		assert.Contains(t, tag1, "createdAt")
		assert.Contains(t, tag1, "updatedAt")
	})
}
