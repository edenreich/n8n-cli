// Client is a simple client for interacting with n8n API
package n8n

// ClientInterface defines the contract for client objects
//
//go:generate go tool counterfeiter -o clientfakes/fake_client.go . ClientInterface
type ClientInterface interface {
	// GetWorkflows fetches workflows from the n8n API
	GetWorkflows() (*WorkflowList, error)
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
	// GetExecutions fetches workflow executions from the n8n API
	GetExecutions(workflowID string, includeData bool, status string, limit int, cursor string) (*ExecutionList, error)
	// GetExecutionById fetches a specific execution by its ID
	GetExecutionById(executionID string, includeData bool) (*Execution, error)
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)
