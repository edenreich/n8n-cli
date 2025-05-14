// Client is a simple client for interacting with n8n API
package n8n

// ClientInterface defines the contract for client objects
//
//go:generate counterfeiter -o clientfakes/fake_client.go . ClientInterface
type ClientInterface interface {
	// GetWorkflows fetches workflows from the n8n API
	GetWorkflows() (*WorkflowList, error)
	// ActivateWorkflow activates a workflow by its ID
	ActivateWorkflow(id string) (*Workflow, error)
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)
