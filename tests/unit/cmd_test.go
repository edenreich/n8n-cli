// Package unit contains unit tests for the n8n-cli
package unit

import (
	"bytes"
	"testing"

	"github.com/edenreich/n8n-cli/cmd"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Simple Name", "Simple_Name"},
		{"Name with spaces", "Name_with_spaces"},
		{"Name/With/Slashes", "Name_With_Slashes"},
		{"Name.With.Dots", "Name.With.Dots"},
		{"Name-With-Dashes", "Name-With-Dashes"},
		{"Name_With_Underscores", "Name_With_Underscores"},
		{"Name With Special Chars: $%^&*", "Name_With_Special_Chars_______"},
		{"Name With Emojis 😀👍", "Name_With_Emojis___"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := cmd.SanitizeFilename(tc.input)
			assert.Equal(t, tc.expected, result, "Expected sanitized filename to match")
		})
	}
}

func TestFormatAPIBaseURL(t *testing.T) {
	testCases := []struct {
		name            string
		instanceURL     string
		expectedBaseURL string
	}{
		{
			name:            "URL with trailing slash",
			instanceURL:     "http://localhost:5678/",
			expectedBaseURL: "http://localhost:5678/api/v1",
		},
		{
			name:            "URL without trailing slash",
			instanceURL:     "http://localhost:5678",
			expectedBaseURL: "http://localhost:5678/api/v1",
		},
		{
			name:            "URL with path",
			instanceURL:     "http://localhost:5678/n8n",
			expectedBaseURL: "http://localhost:5678/n8n/api/v1",
		},
		{
			name:            "URL already with api/v1",
			instanceURL:     "http://localhost:5678/api/v1",
			expectedBaseURL: "http://localhost:5678/api/v1",
		},
		{
			name:            "URL with api/v1 and trailing slash",
			instanceURL:     "http://localhost:5678/api/v1/",
			expectedBaseURL: "http://localhost:5678/api/v1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cmd.FormatAPIBaseURL(tc.instanceURL)
			assert.Equal(t, tc.expectedBaseURL, result, "Expected correctly formatted API base URL")
		})
	}
}

func TestVersionCommand(t *testing.T) {
	origVersion := cmd.Version
	origBuildDate := cmd.BuildDate
	origCommit := cmd.Commit

	cmd.Version = "1.2.3"
	cmd.BuildDate = "2025-05-13"
	cmd.Commit = "abcdef123456"

	buf := new(bytes.Buffer)
	versionCmd := cmd.GetVersionCmd()
	versionCmd.SetOut(buf)

	err := versionCmd.RunE(versionCmd, []string{})
	assert.NoError(t, err, "Expected no error when executing version command")

	output := buf.String()
	assert.Contains(t, output, "n8n-cli 1.2.3", "Version should be included in output")
	assert.Contains(t, output, "Build Date: 2025-05-13", "Build date should be included in output")
	assert.Contains(t, output, "Git Commit: abcdef123456", "Commit should be included in output")

	cmd.Version = origVersion
	cmd.BuildDate = origBuildDate
	cmd.Commit = origCommit
}
