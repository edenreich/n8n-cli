// Client is a simple client for interacting with n8n API
package n8n

// ClientInterface defines the contract for client objects
//
//go:generate go tool counterfeiter -o clientfakes/fake_client.go . ClientInterface
type ClientInterface interface {
	// GetWorkflows fetches workflows from the n8n API
	GetWorkflows() (*WorkflowList, error)
	// GetWorkflow fetches a single workflow by its ID
	GetWorkflow(id string) (*Workflow, error)
	// ActivateWorkflow activates a workflow by its ID
	ActivateWorkflow(id string) (*Workflow, error)
	// DeactivateWorkflow deactivates a workflow by its ID
	DeactivateWorkflow(id string) (*Workflow, error)
	// CreateWorkflow creates a new workflow
	CreateWorkflow(workflow *Workflow) (*Workflow, error)
	// UpdateWorkflow updates an existing workflow by its ID
	UpdateWorkflow(id string, workflow *Workflow) (*Workflow, error)
	// DeleteWorkflow deletes a workflow by its ID
	DeleteWorkflow(id string) error
	// GetWorkflowTags fetches the tags of a workflow by its ID
	GetWorkflowTags(id string) (WorkflowTags, error)
	// UpdateWorkflowTags updates the tags of a workflow by its ID
	UpdateWorkflowTags(id string, tagIds TagIds) (WorkflowTags, error)
	// CreateTag creates a new tag in n8n
	CreateTag(tagName string) (*Tag, error)
	// GetTags fetches all tags from n8n
	GetTags() (*TagList, error)
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)
