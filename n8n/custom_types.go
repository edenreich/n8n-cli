package n8n

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// A flexible ID type that can handle both string and numeric values from the API
type FlexibleID struct {
	Value string
}

func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	var rawValue interface{}
	if err := json.Unmarshal(data, &rawValue); err != nil {
		return err
	}

	switch v := rawValue.(type) {
	case string:
		f.Value = v
	case float64:
		f.Value = strconv.FormatFloat(v, 'f', 0, 64)
	case int:
		f.Value = strconv.Itoa(v)
	case int64:
		f.Value = strconv.FormatInt(v, 10)
	default:
		return fmt.Errorf("unsupported type for ID: %T", v)
	}

	return nil
}

// ExecutionWithFlexibleIDs is a version of Execution with flexible ID fields
type ExecutionWithFlexibleIDs struct {
	CustomData     *map[string]interface{} `json:"customData,omitempty"`
	Data           *map[string]interface{} `json:"data,omitempty"`
	Finished       *bool                   `json:"finished,omitempty"`
	Id             *FlexibleID             `json:"id,omitempty"`
	Mode           *ExecutionMode          `json:"mode,omitempty"`
	RetryOf        *FlexibleID             `json:"retryOf,omitempty"`
	RetrySuccessId *FlexibleID             `json:"retrySuccessId,omitempty"`
	StartedAt      *time.Time              `json:"startedAt,omitempty"`
	StoppedAt      *time.Time              `json:"stoppedAt,omitempty"`
	WaitTill       *time.Time              `json:"waitTill,omitempty"`
	WorkflowId     *FlexibleID             `json:"workflowId,omitempty"`
}

// ExecutionListWithFlexibleIDs is a version of ExecutionList with flexible ID fields
type ExecutionListWithFlexibleIDs struct {
	Data       *[]ExecutionWithFlexibleIDs `json:"data,omitempty"`
	NextCursor *string                     `json:"nextCursor,omitempty"`
}

// Convert to standard ExecutionList type
func (e *ExecutionListWithFlexibleIDs) ToExecutionList() *ExecutionList {
	if e == nil {
		return nil
	}

	var executions []Execution
	if e.Data != nil {
		executions = make([]Execution, len(*e.Data))
		for i, item := range *e.Data {
			executions[i] = toExecution(item)
		}
	}

	return &ExecutionList{
		Data:       &executions,
		NextCursor: e.NextCursor,
	}
}

// Convert to standard Execution type
func toExecution(e ExecutionWithFlexibleIDs) Execution {
	result := Execution{
		CustomData: e.CustomData,
		Data:       e.Data,
		Finished:   e.Finished,
		Mode:       e.Mode,
		StartedAt:  e.StartedAt,
		StoppedAt:  e.StoppedAt,
		WaitTill:   e.WaitTill,
	}

	// Convert ID fields from string to float32
	if e.Id != nil {
		val, err := strconv.ParseFloat(e.Id.Value, 32)
		if err == nil {
			id := float32(val)
			result.Id = &id
		}
	}

	if e.WorkflowId != nil {
		val, err := strconv.ParseFloat(e.WorkflowId.Value, 32)
		if err == nil {
			id := float32(val)
			result.WorkflowId = &id
		}
	}

	if e.RetryOf != nil {
		val, err := strconv.ParseFloat(e.RetryOf.Value, 32)
		if err == nil {
			id := float32(val)
			result.RetryOf = &id
		}
	}

	if e.RetrySuccessId != nil {
		val, err := strconv.ParseFloat(e.RetrySuccessId.Value, 32)
		if err == nil {
			id := float32(val)
			result.RetrySuccessId = &id
		}
	}

	return result
}
