// Package unit contains unit tests for the n8n-cli
package unit

import (
	"time"
)

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func timePtr(s string) *time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return &t
}
