package unit

import (
	"testing"

	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectWorkflowChanges_StaticData(t *testing.T) {
	compactBytes := []byte(`{"lastRun":1234567890,"counter":42}`)
	indentedBytes := []byte("{\n  \"lastRun\": 1234567890,\n  \"counter\": 42\n}")

	compactStaticData := &n8n.Workflow_StaticData{}
	err := compactStaticData.UnmarshalJSON(compactBytes)
	require.NoError(t, err)

	indentedStaticData := &n8n.Workflow_StaticData{}
	err = indentedStaticData.UnmarshalJSON(indentedBytes)
	require.NoError(t, err)

	remote := &n8n.Workflow{
		Name:        "Test Workflow",
		Id:          stringPtr("abc123"),
		Active:      boolPtr(true),
		StaticData:  compactStaticData,
		Connections: make(map[string]interface{}),
	}

	local := &n8n.Workflow{
		Name:        "Test Workflow",
		Id:          stringPtr("abc123"),
		Active:      boolPtr(true),
		StaticData:  indentedStaticData,
		Connections: make(map[string]interface{}),
	}

	changes := workflows.DetectWorkflowChanges(local, remote)

	assert.False(t, changes.NeedsUpdate,
		"NeedsUpdate should be false when only staticData byte formatting differs between local and remote")
}
