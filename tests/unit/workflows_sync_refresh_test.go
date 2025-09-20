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
