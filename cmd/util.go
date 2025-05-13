// Package cmd contains commands for the n8n-cli
package cmd

import (
	"fmt"
	"strings"

	"github.com/edenreich/n8n-cli/n8n"
)

// FormatAPIBaseURL ensures the base URL ends with /api/v1
func FormatAPIBaseURL(instanceURL string) string {
	// Remove trailing slash if present
	if strings.HasSuffix(instanceURL, "/") {
		instanceURL = instanceURL[:len(instanceURL)-1]
	}

	// Add api/v1 path if not already present
	if !strings.HasSuffix(instanceURL, "/api/v1") {
		instanceURL = instanceURL + "/api/v1"
	}

	return instanceURL
}

// FindWorkflow looks up a workflow by exact name match in a list of workflows
// This is a simplified version for unit testing purposes
func FindWorkflow(name string, workflows []n8n.Workflow) (string, error) {
	for _, wf := range workflows {
		if wf.Name == name {
			return *wf.Id, nil
		}
	}

	return "", fmt.Errorf("workflow with name '%s' not found", name)
}
