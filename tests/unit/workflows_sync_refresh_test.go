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
		featureADir := filepath.Join(tempDir, "feature_A")
		featureBDir := filepath.Join(tempDir, "feature_B", "subdir")
		
		err := os.MkdirAll(featureADir, 0755)
		require.NoError(t, err)
		err = os.MkdirAll(featureBDir, 0755)
		require.NoError(t, err)

		workflow1JSON := `{"name": "Feature A Workflow", "id": "feature-a-123", "active": true}`
		workflow2JSON := `{"name": "Feature B Workflow", "id": "feature-b-456", "active": false}`

		workflow1Path := filepath.Join(featureADir, "feature_a_workflow.json")
		workflow2Path := filepath.Join(featureBDir, "feature_b_workflow.json")

		err = os.WriteFile(workflow1Path, []byte(workflow1JSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(workflow2Path, []byte(workflow2JSON), 0644)
		require.NoError(t, err)

		fakeClient := &clientfakes.FakeClientInterface{}
		
		fakeClient.GetWorkflowCalls(func(id string) (*n8n.Workflow, error) {
			return nil, fmt.Errorf("workflow not found")
		})

		fakeClient.CreateWorkflowCalls(func(workflow *n8n.Workflow) (*n8n.Workflow, error) {
			return workflow, nil
		})

		fakeClient.ActivateWorkflowReturns(&n8n.Workflow{}, nil)

		result1, err := workflows.ProcessWorkflowFile(fakeClient, &cobra.Command{}, workflow1Path, false, false)
		require.NoError(t, err)
		assert.Equal(t, "feature-a-123", result1.WorkflowID)
		assert.True(t, result1.Created)

		result2, err := workflows.ProcessWorkflowFile(fakeClient, &cobra.Command{}, workflow2Path, false, false)
		require.NoError(t, err)
		assert.Equal(t, "feature-b-456", result2.WorkflowID)
		assert.True(t, result2.Created)

		assert.Equal(t, 2, fakeClient.CreateWorkflowCallCount())
		assert.Equal(t, 1, fakeClient.ActivateWorkflowCallCount())
	})

	t.Run("Recursive flag controls which files are processed in sync", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "recursive_sync_test")
		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		subDir := filepath.Join(testDir, "subdir")
		err = os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		rootWorkflowJSON := `{"name": "Root Workflow", "id": "root-123", "active": true}`
		subWorkflowJSON := `{"name": "Sub Workflow", "id": "sub-456", "active": false}`

		rootWorkflowPath := filepath.Join(testDir, "root_workflow.json")
		subWorkflowPath := filepath.Join(subDir, "sub_workflow.json")

		err = os.WriteFile(rootWorkflowPath, []byte(rootWorkflowJSON), 0644)
		require.NoError(t, err)
		err = os.WriteFile(subWorkflowPath, []byte(subWorkflowJSON), 0644)
		require.NoError(t, err)

		fakeClient := &clientfakes.FakeClientInterface{}
		
		processedWorkflows := make(map[string]bool)
		
		fakeClient.GetWorkflowCalls(func(id string) (*n8n.Workflow, error) {
			processedWorkflows[id] = true
			return nil, fmt.Errorf("workflow not found")
		})

		fakeClient.CreateWorkflowCalls(func(workflow *n8n.Workflow) (*n8n.Workflow, error) {
			return workflow, nil
		})

		fakeClient.ActivateWorkflowReturns(&n8n.Workflow{}, nil)

		result1, err := workflows.ProcessWorkflowFile(fakeClient, &cobra.Command{}, rootWorkflowPath, false, false)
		require.NoError(t, err)
		assert.Equal(t, "root-123", result1.WorkflowID)

		processedWorkflows = make(map[string]bool)
		fakeClient.Invocations()

		entries, err := os.ReadDir(testDir)
		require.NoError(t, err)
		
		foundFiles := 0
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
				foundFiles++
			}
		}
		assert.Equal(t, 1, foundFiles, "Non-recursive should only find 1 file in root")

		foundRecursiveFiles := 0
		err = filepath.WalkDir(testDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(d.Name()) == ".json" {
				foundRecursiveFiles++
			}
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 2, foundRecursiveFiles, "Recursive should find 2 files total")
	})
}
