package unit

import (
	"testing"

	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDetectWorkflowChanges_Tags(t *testing.T) {
	tests := []struct {
		name     string
		local    *n8n.Workflow
		remote   *n8n.Workflow
		wantTags bool
	}{
		{
			name: "Local has tags, remote doesn't",
			local: &n8n.Workflow{
				Name: "Test Workflow",
				Tags: &n8n.WorkflowTags{
					{Id: stringPtr("1"), Name: "tag1"},
				},
			},
			remote: &n8n.Workflow{
				Name: "Test Workflow",
				Tags: &n8n.WorkflowTags{},
			},
			wantTags: true,
		},
		{
			name: "Local and remote have different tags",
			local: &n8n.Workflow{
				Name: "Test Workflow",
				Tags: &n8n.WorkflowTags{
					{Id: stringPtr("1"), Name: "tag1"},
					{Id: stringPtr("2"), Name: "tag2"},
				},
			},
			remote: &n8n.Workflow{
				Name: "Test Workflow",
				Tags: &n8n.WorkflowTags{
					{Id: stringPtr("1"), Name: "tag1"},
				},
			},
			wantTags: true,
		},
		{
			name: "Local and remote have same tags",
			local: &n8n.Workflow{
				Name: "Test Workflow",
				Tags: &n8n.WorkflowTags{
					{Id: stringPtr("1"), Name: "tag1"},
				},
			},
			remote: &n8n.Workflow{
				Name: "Test Workflow",
				Tags: &n8n.WorkflowTags{
					{Id: stringPtr("1"), Name: "tag1"},
				},
			},
			wantTags: false,
		},
		{
			name: "Local has no tags, remote does",
			local: &n8n.Workflow{
				Name: "Test Workflow",
				Tags: &n8n.WorkflowTags{},
			},
			remote: &n8n.Workflow{
				Name: "Test Workflow",
				Tags: &n8n.WorkflowTags{
					{Id: stringPtr("1"), Name: "tag1"},
				},
			},
			wantTags: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := workflows.DetectWorkflowChanges(tt.local, tt.remote)
			assert.Equal(t, tt.wantTags, changes.NeedsTagsUpdate)
		})
	}
}

func TestHandleTagUpdates(t *testing.T) {
	fakeClient := &clientfakes.FakeClientInterface{}

	workflow := &n8n.Workflow{
		Name: "Test Workflow",
		Id:   stringPtr("123"),
		Tags: &n8n.WorkflowTags{
			{Id: stringPtr("1"), Name: "tag1"},
			{Id: stringPtr("2"), Name: "tag2"},
		},
	}

	cmd := &cobra.Command{}

	fakeClient.UpdateWorkflowTagsReturns(n8n.WorkflowTags{}, nil)

	err := workflows.HandleTagUpdates(fakeClient, cmd, workflow, *workflow.Id, false)
	assert.NoError(t, err)

	assert.Equal(t, 1, fakeClient.UpdateWorkflowTagsCallCount())
	id, tags := fakeClient.UpdateWorkflowTagsArgsForCall(0)
	assert.Equal(t, "123", id)
	assert.Len(t, tags, 2)
	assert.Equal(t, "1", tags[0].Id)
	assert.Equal(t, "2", tags[1].Id)
}
