// Package integration contains integration tests for the n8n-cli
package integration

import (
	"bytes"
	"testing"

	"github.com/edenreich/n8n-cli/tests"
	"github.com/spf13/cobra"
)

// executeCommand is a helper to execute a command and capture its output
func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	tests.SkipIfNotIntegration(t)

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}
