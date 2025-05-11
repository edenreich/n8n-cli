package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	cmd := GetRootCmd()

	stdout, _, err := executeCommand(cmd)

	if err != nil {
		t.Errorf("Error executing root command: %v", err)
	}

	if !strings.Contains(stdout, "n8n-cli") {
		t.Errorf("Root command help not displayed: %s", stdout)
	}
}

// Helper function to execute commands and capture output
func executeCommand(cmd *cobra.Command, args ...string) (string, string, error) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = append([]string{"n8n-cli"}, args...)

	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}

	oldOut := cmd.OutOrStdout()
	oldErr := cmd.ErrOrStderr()

	cmd.SetOut(stdOut)
	cmd.SetErr(stdErr)

	defer func() {
		cmd.SetOut(oldOut)
		cmd.SetErr(oldErr)
	}()

	err := cmd.Execute()

	return stdOut.String(), stdErr.String(), err
}
