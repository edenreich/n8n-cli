// Package unit contains unit tests for the n8n-cli
package unit

import (
	"errors"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/stretchr/testify/assert"
)

func TestGetWorkflow(t *testing.T) {
	testCases := []struct {
		name          string
		workflowID    string
		mockReturnWF  *n8n.Workflow
		mockReturnErr error
		expectError   bool
	}{
		{
			name:          "Successfully gets workflow",
			workflowID:    "123",
			mockReturnWF:  &n8n.Workflow{Name: "Test Workflow"},
			mockReturnErr: nil,
			expectError:   false,
		},
		{
			name:          "Returns error when API call fails",
			workflowID:    "789",
			mockReturnWF:  nil,
			mockReturnErr: errors.New("API error"),
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := &clientfakes.FakeClientInterface{}
			fakeClient.GetWorkflowReturns(tc.mockReturnWF, tc.mockReturnErr)

			workflow, err := fakeClient.GetWorkflow(tc.workflowID)

			// Check that the correct workflow ID was passed
			id := fakeClient.GetWorkflowArgsForCall(0)
			assert.Equal(t, tc.workflowID, id)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.mockReturnErr, err)
				assert.Nil(t, workflow)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockReturnWF, workflow)
			}
		})
	}
}
