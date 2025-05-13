// Package integration contains integration tests for the n8n-cli
package integration

import (
	"strings"
	"testing"

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/tests"
	"github.com/stretchr/testify/assert"
)

// TestRootCommand tests the root command of the CLI
func TestRootCommand(t *testing.T) {
	tests.SkipIfNotIntegration(t)

	rootCmd := cmd.GetRootCmd()

	stdout, _, err := executeCommand(t, rootCmd)

	assert.NoError(t, err, "Error executing root command")
	assert.Contains(t, stdout, "n8n-cli", "Root command help not displayed")
}

// TestVersionCommand tests the version command and flags
func TestVersionCommand(t *testing.T) {
	tests.SkipIfNotIntegration(t)

	rootCmd := cmd.GetRootCmd()

	expectedLines := []string{
		"n8n-cli",
		"Build Date:",
		"Git Commit:",
	}

	testCases := []struct {
		name string
		args []string
	}{
		{"Long Flag", []string{"--version"}},
		{"Short Flag", []string{"-v"}},
		{"Subcommand", []string{"version"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, _, err := executeCommand(t, rootCmd, tc.args...)

			assert.NoError(t, err, "Error executing command with %v", tc.args)

			for _, expected := range expectedLines {
				assert.True(t, strings.Contains(stdout, expected),
					"Expected output to contain '%s', but got: %s", expected, stdout)
			}
		})
	}
}
