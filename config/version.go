package config

// Version information set by semantic release
// These variables can be overridden during build time using ldflags

var (
	// Version holds the current version of n8n-cli
	// This will be replaced during semantic release with the actual version
	Version = "dev"

	// BuildDate holds the date when the binary was built
	// This will be replaced during build with the actual build date
	BuildDate = "unknown"

	// Commit holds the git commit hash used to build the binary
	// This will be replaced during build with the actual git commit hash
	Commit = "none"
)
