// Package unit contains unit tests for the n8n-cli
package unit

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestActivateCommandExecute(t *testing.T) {
	testCases := []struct {
		name           string
		workflowID     string
		mockReturnWF   *n8n.Workflow
		mockReturnErr  error
		expectedOutput string
		expectError    bool
	}{
		{
			name:          "Successfully activates workflow",
			workflowID:    "123",
			mockReturnWF:  &n8n.Workflow{Name: "Test Workflow"},
			mockReturnErr: nil,
			expectedOutput: "Workflow with ID 123 has been activated successfully\n" +
				"Name: Test Workflow\n",
			expectError: false,
		},
		{
			name:           "Successfully activates workflow without name",
			workflowID:     "456",
			mockReturnWF:   &n8n.Workflow{},
			mockReturnErr:  nil,
			expectedOutput: "Workflow with ID 456 has been activated successfully\n",
			expectError:    false,
		},
		{
			name:           "Returns error when API call fails",
			workflowID:     "789",
			mockReturnWF:   nil,
			mockReturnErr:  errors.New("API error"),
			expectedOutput: "Error activating workflow: API error\n",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := &clientfakes.FakeClientInterface{}

			fakeClient.ActivateWorkflowReturns(tc.mockReturnWF, tc.mockReturnErr)

			outBuf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)

			cmd := &cobra.Command{}
			cmd.SetOut(outBuf)
			cmd.SetErr(errBuf)

			origRunE := workflows.ActivateCmd.RunE
			defer func() {
				workflows.ActivateCmd.RunE = origRunE
			}()

			testRunE := func(cmd *cobra.Command, args []string) error {
				if len(args) == 0 {
					return errors.New("this command requires a workflow ID")
				}

				workflowID := args[0]
				workflow, err := fakeClient.ActivateWorkflow(workflowID)

				if err != nil {
					cmd.PrintErrf("Error activating workflow: %v\n", err)
					return err
				}

				cmd.Printf("Workflow with ID %s has been activated successfully\n", workflowID)
				if workflow != nil && workflow.Name != "" {
					cmd.Printf("Name: %s\n", workflow.Name)
				}

				return nil
			}

			workflows.ActivateCmd.RunE = testRunE

			err := workflows.ActivateCmd.RunE(cmd, []string{tc.workflowID})

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			output := outBuf.String() + errBuf.String()
			assert.Equal(t, tc.expectedOutput, output)

			assert.Equal(t, 1, fakeClient.ActivateWorkflowCallCount())
			passedWorkflowID := fakeClient.ActivateWorkflowArgsForCall(0)
			assert.Equal(t, tc.workflowID, passedWorkflowID)
		})
	}
}

func TestActivateCmd_NoArgs(t *testing.T) {
	cmd := &cobra.Command{}

	runE := workflows.ActivateCmd.RunE

	err := runE(cmd, []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "this command requires a workflow ID")
}
