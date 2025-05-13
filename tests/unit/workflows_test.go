// Package unit contains unit tests for the n8n-cli
package unit

import (
	"testing"

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/stretchr/testify/assert"
)

func TestFindWorkflow(t *testing.T) {
	// Setup test workflows
	workflows := []struct {
		Id     *string
		Name   string
		Active *bool
	}{
		{Id: stringPtr("1"), Name: "First Workflow", Active: boolPtr(true)},
		{Id: stringPtr("2"), Name: "Second Workflow", Active: boolPtr(false)},
		{Id: stringPtr("3"), Name: "Third Workflow", Active: boolPtr(true)},
	}

	testCases := []struct {
		name           string
		searchName     string
		expectedID     string
		expectNotFound bool
	}{
		{
			name:           "Exact match",
			searchName:     "Second Workflow",
			expectedID:     "2",
			expectNotFound: false,
		},
		{
			name:           "No match",
			searchName:     "Missing Workflow",
			expectedID:     "",
			expectNotFound: true,
		},
		{
			name:           "Partial match should fail",
			searchName:     "Workflow",
			expectedID:     "",
			expectNotFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert to format expected by FindWorkflow
			var testWorkflows []n8n.Workflow
			for _, wf := range workflows {
				wfMap := n8n.Workflow{
					Id:     wf.Id,
					Name:   wf.Name,
					Active: wf.Active,
				}
				testWorkflows = append(testWorkflows, wfMap)
			}

			id, err := cmd.FindWorkflow(tc.searchName, testWorkflows)

			if tc.expectNotFound {
				assert.Error(t, err, "Expected workflow not to be found")
				assert.Empty(t, id, "Expected empty ID when workflow not found")
			} else {
				assert.NoError(t, err, "Expected to find workflow")
				assert.Equal(t, tc.expectedID, id, "Expected correct workflow ID")
			}
		})
	}
}

// Using helper functions from unit package
