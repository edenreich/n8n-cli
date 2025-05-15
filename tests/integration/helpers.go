// Package integration contains integration tests for the n8n-cli
package integration

import (
	"bytes"
	"testing"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/spf13/cobra"
)

// executeCommand is a helper to execute a command and capture its output
func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	// Create a fresh root command for each test to avoid flag persistence issues
	rootCmd := rootcmd.GetRootCmd()

	// Build the full command path
	var cmdArgs []string
	if cmd.Use == "list" && cmd.Parent() != nil && cmd.Parent().Use == "workflows" {
		cmdArgs = append([]string{"workflows", "list"}, args...)
	} else if cmd != rootCmd {
		cmdPath := []string{}
		current := cmd

		for current != nil && current != rootCmd {
			cmdPath = append([]string{current.Use}, cmdPath...)
			current = current.Parent()
		}

		cmdArgs = append(cmdPath, args...)
	} else {
		cmdArgs = args
	}

	// Debug the command being run
	t.Logf("Running command: %v", cmdArgs)

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(cmdArgs)
	err := rootCmd.Execute()
	return stdout.String(), stderr.String(), err
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
