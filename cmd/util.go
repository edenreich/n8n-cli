// Package cmd contains commands for the n8n-cli
package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/edenreich/n8n-cli/n8n"
)

// FormatAPIBaseURL ensures the base URL ends with /api/v1
func FormatAPIBaseURL(instanceURL string) string {
	instanceURL = strings.TrimSuffix(instanceURL, "/")

	if !strings.HasSuffix(instanceURL, "/api/v1") {
		instanceURL = instanceURL + "/api/v1"
	}

	return instanceURL
}

// FindWorkflow looks up a workflow by exact name match in a list of workflows
func FindWorkflow(name string, workflows []n8n.Workflow) (string, error) {
	for _, wf := range workflows {
		if wf.Name == name {
			return *wf.Id, nil
		}
	}

	return "", fmt.Errorf("workflow with name '%s' not found", name)
}

// SanitizeFilename converts a workflow name to a valid filename
func SanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	name = strings.ReplaceAll(name, "$", "_")
	name = strings.ReplaceAll(name, "%", "_")
	name = strings.ReplaceAll(name, "^", "_")
	name = strings.ReplaceAll(name, "&", "_")

	var result strings.Builder
	for _, r := range name {
		if r >= 0x1F000 {
			result.WriteRune('_')
		} else {
			result.WriteRune(r)
		}
	}

	if result.Len() > 0 {
		return result.String()
	}

	return name
}

// DetectWorkflowDrift compares two workflows and returns true if they differ
// This function uses reflect.DeepEqual for accurate structural comparison
// If minimal is true, both workflows will be cleaned before comparison
func DetectWorkflowDrift(actual n8n.Workflow, desired n8n.Workflow, minimal bool) bool {
	if minimal {
		actual = n8n.CleanWorkflow(actual)
		desired = n8n.CleanWorkflow(desired)
	}

	return !reflect.DeepEqual(actual, desired)
}
