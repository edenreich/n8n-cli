package unit

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExecutionsCommand(t *testing.T) {
	fakeClient := &clientfakes.FakeClientInterface{}
	var stdout, stderr bytes.Buffer

	setupTestCommand := func() *cobra.Command {
		cmd := &cobra.Command{
			Use: "executions [WORKFLOW_ID]",
			RunE: func(cmd *cobra.Command, args []string) error {
				return workflows.ExecutionHandler{Client: fakeClient}.Handle(cmd, args)
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().Bool("include-data", false, "")
		cmd.Flags().String("status", "", "")
		cmd.Flags().Int("limit", 10, "")
		cmd.Flags().String("cursor", "", "")
		cmd.Flags().Bool("json", false, "")
		cmd.Flags().Bool("raw", false, "")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	createSampleExecutionList := func(count int) *n8n.ExecutionList {
		executions := make([]n8n.Execution, count)

		now := time.Now()
		finished := true
		mode := n8n.ExecutionModeManual

		for i := 0; i < count; i++ {
			id := float32(1000 + i)
			workflowId := float32(100 + i)
			startedAt := now.Add(time.Duration(-i) * time.Hour)
			stoppedAt := startedAt.Add(30 * time.Second)

			executions[i] = n8n.Execution{
				Id:         &id,
				WorkflowId: &workflowId,
				Finished:   &finished,
				Mode:       &mode,
				StartedAt:  &startedAt,
				StoppedAt:  &stoppedAt,
			}
		}

		data := executions
		nextCursor := "next-page-cursor"

		return &n8n.ExecutionList{
			Data:       &data,
			NextCursor: &nextCursor,
		}
	}

	t.Run("successfully gets executions", func(t *testing.T) {
		cmd := setupTestCommand()
		executions := createSampleExecutionList(3)
		fakeClient.GetExecutionsReturns(executions, nil)

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Execution history")
		assert.Contains(t, stdout.String(), "next-page-cursor")

		_, includeData, status, limit, cursor := fakeClient.GetExecutionsArgsForCall(0)
		assert.False(t, includeData)
		assert.Empty(t, status)
		assert.Equal(t, 10, limit)
		assert.Empty(t, cursor)
	})

	t.Run("returns error when client fails", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetExecutionsReturns(nil, errors.New("API error"))

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, stderr.String(), "Error getting executions")
	})

	t.Run("shows empty message when no executions", func(t *testing.T) {
		cmd := setupTestCommand()
		emptyData := []n8n.Execution{}
		executions := &n8n.ExecutionList{
			Data: &emptyData,
		}
		fakeClient.GetExecutionsReturns(executions, nil)

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "No executions found")
	})

	t.Run("outputs JSON when requested", func(t *testing.T) {
		cmd := setupTestCommand()
		err := cmd.Flags().Set("json", "true")
		assert.NoError(t, err, "Failed to set json flag")
		executions := createSampleExecutionList(1)
		fakeClient.GetExecutionsReturns(executions, nil)

		err = cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "\"data\":")
		assert.Contains(t, stdout.String(), "\"nextCursor\":")
	})
}
