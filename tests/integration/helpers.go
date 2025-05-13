// Package integration contains integration tests for the n8n-cli
package integration

import (
	"bytes"
	"testing"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/tests"
	"github.com/spf13/cobra"
)

// executeCommand is a helper to execute a command and capture its output
func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	tests.SkipIfNotIntegration(t)

	rootCmd := rootcmd.GetRootCmd()
	if cmd.Use == "list" && cmd.Parent() != nil && cmd.Parent().Use == "workflows" {
		args = append([]string{"workflows", "list"}, args...)
	} else if cmd != rootCmd {
		cmdPath := []string{}
		current := cmd

		for current != nil && current != rootCmd {
			cmdPath = append([]string{current.Use}, cmdPath...)
			current = current.Parent()
		}

		args = append(cmdPath, args...)
	}

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return stdout.String(), stderr.String(), err
}
