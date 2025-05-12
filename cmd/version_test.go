package cmd

import (
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := rootCmd

	expectedLines := []string{
		"n8n-cli dev",
		"Build Date: unknown",
		"Git Commit: none",
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
			stdout, _, err := executeCommand(cmd, tc.args...)

			if err != nil {
				t.Errorf("Error executing command with %v: %v", tc.args, err)
			}

			for _, line := range expectedLines {
				if !strings.Contains(stdout, line) {
					t.Logf("Actual output: %q", stdout)
					t.Errorf("Expected output to contain %q", line)
				}
			}
		})
	}
}
