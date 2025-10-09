package unit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessWorkflowFile_ReturnsWorkflowResult(t *testing.T) {
	fakeClient := &clientfakes.FakeClientInterface{}
	cmd := &cobra.Command{}

	tempDir, err := os.MkdirTemp("", "n8n-cli-test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}()

	testFilePath := filepath.Join(tempDir, "test-workflow.json")

	workflowID := "test-id-123"
	workflow := n8n.Workflow{
		Name: "Test Workflow",
		Id:   &workflowID,
	}

	fakeClient.GetWorkflowReturns(nil, fmt.Errorf("workflow not found"))

	newID := "new-id-456"
	fakeClient.CreateWorkflowReturns(&n8n.Workflow{
		Id:   &newID,
		Name: workflow.Name,
	}, nil)

	workflowJSON := `{"name": "Test Workflow", "id": "test-id-123"}`
	err = os.WriteFile(testFilePath, []byte(workflowJSON), 0644)
	require.NoError(t, err)

	result, err := workflows.ProcessWorkflowFile(fakeClient, cmd, testFilePath, false, false)
	require.NoError(t, err)

	assert.Equal(t, newID, result.WorkflowID)
	assert.Equal(t, "Test Workflow", result.Name)
	assert.Equal(t, testFilePath, result.FilePath)
	assert.True(t, result.Created)
	assert.False(t, result.Updated)
}

func TestWorkflowResult(t *testing.T) {
	result := workflows.WorkflowResult{
		WorkflowID: "123",
		Name:       "Test Workflow",
		FilePath:   "/path/to/file",
		Created:    true,
		Updated:    false,
	}

	assert.Equal(t, "123", result.WorkflowID)
	assert.Equal(t, "Test Workflow", result.Name)
	assert.Equal(t, "/path/to/file", result.FilePath)
	assert.True(t, result.Created)
	assert.False(t, result.Updated)
}

func TestSyncNestedDirectories(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Processes workflow files from nested directories", func(t *testing.T) {
		// Create nested directory structure
		featureADir := filepath.Join(tempDir, "feature_A")
		featureBDir := filepath.Join(tempDir, "feature_B", "subdir")
		
		err := os.MkdirAll(featureADir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(featureBDir, 0755)
		require.NoError(t, err)

		// Create workflow files in nested directories
		workflow1JSON := `{"name": "Feature A Workflow", "id": "feature-a-123", "active": true}`
		workflow2JSON := `{"name": "Feature B Workflow", "id": "feature-b-456", "active": false}`

		workflow1Path := filepath.Join(featureADir, "feature_a_workflow.json")
		workflow2Path := filepath.Join(featureBDir, "feature_b_workflow.json")

		err = os.WriteFile(workflow1Path, []byte(workflow1JSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(workflow2Path, []byte(workflow2JSON), 0644)
		require.NoError(t, err)

		// Mock client responses
		fakeClient := &clientfakes.FakeClientInterface{}
		
		// Mock GetWorkflow to return not found (will trigger create)
		fakeClient.GetWorkflowCalls(func(id string) (*n8n.Workflow, error) {
			return nil, fmt.Errorf("workflow not found")
		})

		// Mock CreateWorkflow to return created workflows
		fakeClient.CreateWorkflowCalls(func(workflow *n8n.Workflow) (*n8n.Workflow, error) {
			return workflow, nil
		})

		// Mock ActivateWorkflow
		fakeClient.ActivateWorkflowReturns(&n8n.Workflow{}, nil)

		// Process the nested workflow files
		result1, err := workflows.ProcessWorkflowFile(fakeClient, &cobra.Command{}, workflow1Path, false, false)
		require.NoError(t, err)
		assert.Equal(t, "feature-a-123", result1.WorkflowID)
		assert.True(t, result1.Created)

		result2, err := workflows.ProcessWorkflowFile(fakeClient, &cobra.Command{}, workflow2Path, false, false)
		require.NoError(t, err)
		assert.Equal(t, "feature-b-456", result2.WorkflowID)
		assert.True(t, result2.Created)

		// Verify both CreateWorkflow and ActivateWorkflow were called
		assert.Equal(t, 2, fakeClient.CreateWorkflowCallCount())
		assert.Equal(t, 1, fakeClient.ActivateWorkflowCallCount()) // Only one workflow is active
	})
}
