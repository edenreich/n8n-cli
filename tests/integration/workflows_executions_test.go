// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowExecutions(t *testing.T) {
	now := time.Now()
	finished := true
	mode := n8n.ExecutionModeManual

	executions123 := []n8n.Execution{
		{
			Id:         float32Ptr(1001),
			WorkflowId: float32Ptr(123),
			Finished:   &finished,
			Mode:       &mode,
			StartedAt:  &now,
			StoppedAt:  timePtr(now.Add(30 * time.Second)),
		},
		{
			Id:         float32Ptr(1002),
			WorkflowId: float32Ptr(123),
			Finished:   &finished,
			Mode:       &mode,
			StartedAt:  timePtr(now.Add(-1 * time.Hour)),
			StoppedAt:  timePtr(now.Add(-1*time.Hour + 45*time.Second)),
		},
	}

	executions456 := []n8n.Execution{
		{
			Id:         float32Ptr(2001),
			WorkflowId: float32Ptr(456),
			Finished:   &finished,
			Mode:       &mode,
			StartedAt:  timePtr(now.Add(-2 * time.Hour)),
			StoppedAt:  timePtr(now.Add(-2*time.Hour + 15*time.Second)),
		},
	}

	allExecutions := append(executions123, executions456...)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-N8N-API-KEY")
		if apiKey != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/v1/executions") {
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) > 4 && pathParts[4] != "" {
				executionID := pathParts[4]

				var foundExecution *n8n.Execution
				for _, exec := range allExecutions {
					if fmt.Sprintf("%v", *exec.Id) == executionID {
						foundExecution = &exec
						break
					}
				}

				if foundExecution == nil {
					w.WriteHeader(http.StatusNotFound)
					_, _ = fmt.Fprintln(w, `{"error": "Execution not found"}`)
					return
				}

				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(foundExecution); err != nil {
					http.Error(w, fmt.Sprintf("Error encoding execution: %v", err), http.StatusInternalServerError)
				}
				return
			}

			query := r.URL.Query()
			workflowID := query.Get("workflowId")

			var responseData []n8n.Execution
			switch workflowID {
			case "123":
				responseData = executions123
			case "456":
				responseData = executions456
			default:
				responseData = allExecutions
			}

			response := n8n.ExecutionList{
				Data: &responseData,
			}

			nextCursor := "next-page-cursor"
			response.NextCursor = &nextCursor

			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	setupTestConfig(t, server.URL, "test-api-key")
	defer teardownTestConfig()

	t.Run("Get all executions", func(t *testing.T) {
		output, err := runCommand(t, "workflows", "executions")

		assert.NoError(t, err)
		assert.Contains(t, output, "Execution history for all workflows")
		assert.Contains(t, output, "1001")
		assert.Contains(t, output, "2001")
		assert.Contains(t, output, "next-page-cursor")
	})

	t.Run("Get executions for specific workflow", func(t *testing.T) {
		output, err := runCommand(t, "workflows", "executions", "123")

		assert.NoError(t, err)
		assert.Contains(t, output, "Execution history for workflow ID 123")
		assert.Contains(t, output, "1001")
		assert.NotContains(t, output, "2001")
	})

	t.Run("Get executions with JSON output", func(t *testing.T) {
		output, err := runCommand(t, "workflows", "executions", "--json")

		assert.NoError(t, err)
		assert.Contains(t, output, `"data"`)
		assert.Contains(t, output, `"nextCursor"`)

		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result["data"])
	})

	t.Run("Get executions with limit", func(t *testing.T) {
		output, err := runCommand(t, "workflows", "executions", "--limit", "1")
		assert.NoError(t, err)

		if strings.Contains(output, "{") {
			var result map[string]interface{}
			err := json.Unmarshal([]byte(output), &result)
			assert.NoError(t, err)
			assert.Contains(t, output, `"data"`)
		} else {
			assert.Contains(t, output, "Execution history")
		}
	})

	t.Run("Unauthorized access", func(t *testing.T) {
		setupTestConfig(t, server.URL, "wrong-api-key")

		_, err := runCommand(t, "workflows", "executions")

		assert.Error(t, err)

		setupTestConfig(t, server.URL, "test-api-key")
	})
}
