// Package integration contains integration tests for the n8n-cli
package integration

import (
	"bytes"
	"testing"
	"time"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// executeCommand is a helper to execute a command and capture its output
func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, string, error) {
	rootCmd := rootcmd.GetRootCmd()

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

func float32Ptr(f float32) *float32 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// setupTestConfig configures the viper instance for tests
func setupTestConfig(t *testing.T, instanceURL, apiKey string) {
	viper.Reset()
	viper.Set("instance_url", instanceURL)
	viper.Set("api_key", apiKey)
}

// teardownTestConfig cleans up the test configuration
func teardownTestConfig() {
	viper.Reset()
}

// runCommand runs a command with the given arguments and returns its output
func runCommand(t *testing.T, args ...string) (string, error) {
	rootCmd := rootcmd.GetRootCmd()

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)

	t.Logf("Running command: %v", args)
	err := rootCmd.Execute()

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	return output, err
}
