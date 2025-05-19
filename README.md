<h1 align="center">N8N Command Line Interface (CLI)</h1>

<p align="center">
  <a href="https://github.com/edenreich/n8n-cli/actions/workflows/ci.yml">
    <img src="https://github.com/edenreich/n8n-cli/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
  <a href="https://github.com/edenreich/n8n-cli/releases">
    <img src="https://img.shields.io/github/v/release/edenreich/n8n-cli" alt="Latest Release">
  </a>
  <a href="https://github.com/edenreich/n8n-cli/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/edenreich/n8n-cli" alt="License">
  </a>
  <a href="https://goreportcard.com/report/github.com/edenreich/n8n-cli">
    <img src="https://goreportcard.com/badge/github.com/edenreich/n8n-cli" alt="Go Report Card">
  </a>
  <a href="https://pkg.go.dev/github.com/edenreich/n8n-cli">
    <img src="https://pkg.go.dev/badge/github.com/edenreich/n8n-cli.svg" alt="Go Reference">
  </a>
</p>

<p align="center">Command line interface for managing n8n instances.</p>

## Table of Contents

- [Installation](#installation)
  - [Quick Install](#quick-install-linux-macos-windows-with-wsl)
  - [Autocompletion](#autocompletion)
  - [Manual Installation with Go](#manual-installation-with-go)
- [Configuration](#configuration)
- [Commands](#commands)
  - [Version](#version)
  - [Workflows](#workflows)
    - [List](#list)
    - [Refresh](#refresh)
    - [Sync](#sync)
    - [Activate](#activate)
    - [Deactivate](#deactivate)

## Installation

### Quick Install (Linux, macOS, Windows with WSL)

```bash
curl -sSLf https://raw.github.com/edenreich/n8n-cli/main/install.sh | sh
```

Or install a specific version:

```bash
curl -sSLf https://raw.github.com/edenreich/n8n-cli/main/install.sh | sh -s -- --version v0.1.0-rc.1
```

This script will automatically detect your operating system and architecture and install the appropriate binary.

### Autocompletion

To enable auto completion for `bash`, `zsh`, or `fish`, run the following command:

```bash
source <(n8n completion bash) # for bash
source <(n8n completion zsh)  # for zsh
source <(n8n completion fish) # for fish
```

If you need it permanently, add it to your shell's configuration file (e.g., `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`).

### Manual Installation with Go

```bash
go install github.com/edenreich/n8n-cli@latest
```

## Configuration

Create a `.env` file in your current directory. The CLI will automatically load environment variables from this file.

```
N8N_API_KEY=your_n8n_api_key
N8N_INSTANCE_URL=https://your-instance.n8n.cloud
```

You can generate an API key in the n8n UI under Settings > API.

Alternatively, you can set these environment variables directly in your shell:

```bash
export N8N_API_KEY=your_n8n_api_key
export N8N_INSTANCE_URL=https://your-instance.n8n.cloud
```

Note: Environment variables set directly in your shell will take precedence over those defined in the `.env` file.

**Important:** Never commit your `.env` file containing API credentials to version control systems like GitHub. Make sure to add `.env` to your `.gitignore` file to prevent accidental exposure of sensitive credentials.

## Commands

### Version

Display the version information of the n8n CLI:

```bash
n8n --version
# Or use the explicit command
n8n version
```

### Workflows

Manage n8n workflows with various subcommands.

#### List

List workflows from an n8n instance:

```bash
n8n workflows list
```

Options:

- `--output, -o`: Output format (default: "table"). Supported formats:
  - `table`: Human-readable tabular format
  - `json`: JSON format for programmatic use
  - `yaml`: YAML format for configuration files

Example:

```bash
# List workflows in default table format
n8n workflows list

# List workflows in JSON format
n8n workflows list --output json

# List workflows in YAML format
n8n workflows list --output yaml
```

#### Refresh

Refresh local workflow files with the current state from an n8n instance:

```bash
n8n workflows refresh --directory workflows/
```

The refresh command is an essential step before syncing to ensure you don't accidentally delete or overwrite workflows on the remote n8n instance. It pulls the current state of the workflows from n8n and updates or creates the corresponding local files.

Options:

- `--directory, -d`: Directory to store the workflow files (required)
- `--dry-run`: Show what would be updated without making changes
- `--overwrite`: Overwrite existing files even if they have a different name
- `--output, -o`: Output format for new workflow files (json or yaml)
- `--no-truncate`: Include all fields in output files, including null and optional fields (default: false)
- `--all`: Refresh all workflows from n8n instance, not just those in the directory.

Examples:

```bash
# Refresh only existing workflows in the directory
n8n workflows refresh --directory workflows/

# Refresh all workflows from n8n instance (including new ones)
n8n workflows refresh --directory workflows/ --all

# Preview what would be refreshed without making changes
n8n workflows refresh --directory workflows/ --dry-run

# Refresh workflows and save them as YAML files
n8n workflows refresh --directory workflows/ --output yaml

# Refresh workflows without minimizing the JSON/YAML output
n8n workflows refresh --directory workflows/ --no-truncate
```

#### Sync

Synchronize JSON workflows from a local directory to an n8n instance:

```bash
n8n workflows sync --directory workflows/
```

Options:

- `--directory, -d`: Directory containing workflow JSON/YAML files (required)
- `--dry-run`: Show what would be done without making changes
- `--prune`: Remove workflows from the n8n instance that are not present in the local directory
- `--refresh`: Refresh the local state with the remote state after sync (default: true)
- `--output, -o`: Output format for refreshed workflow files (json or yaml). If not specified, uses the existing file extension in the directory
- `--all`: Refresh all workflows from n8n instance when refreshing, not just those in the directory

How the sync command handles workflow IDs:

1. If a workflow file contains an ID:
   - If that ID exists on the n8n instance, the workflow will be updated
   - If that ID doesn't exist on the n8n instance, a new workflow will be created (n8n API doesn't allow specifying IDs when creating workflows)
2. If a workflow file doesn't have an ID, a new workflow will be created with a server-generated ID

This ensures that workflows maintain their IDs across different environments and prevents duplication.

Example:

```bash
# Sync workflows to the n8n instance
n8n workflows sync --directory workflows/

# Test without making changes
n8n workflows sync --directory workflows/ --dry-run

# Sync workflows and remove any remote workflows not in the local directory
n8n workflows sync --directory workflows/ --prune

# Sync workflows and refresh as JSON (overrides existing format)
n8n workflows sync --directory workflows/ --output json

# Sync workflows and refresh all workflows from n8n instance (including ones not in local directory)
n8n workflows sync --directory workflows/ --all

# Sync workflows without refreshing the local state afterward
n8n workflows sync --directory workflows/ --refresh=false
```

#### Activate

Activate a specific workflow by ID:

```bash
n8n workflows activate WORKFLOW_ID
```

This command activates a workflow in the n8n instance, making it ready to be triggered by events.

#### Deactivate

Deactivate a specific workflow by ID:

```bash
n8n workflows deactivate WORKFLOW_ID
```

This command deactivates a workflow in the n8n instance, stopping it from being triggered by events.

## Development

### Available Tasks

The project uses [Taskfile](https://taskfile.dev) for automating common development operations:

```bash
# Run unit tests
task test-unit

# Run integration tests
task test-integration

# Run all tests
task test-all

# Run linting
task lint

# Build the CLI
task build

# Run the CLI during development (args are passed to the CLI)
task cli -- workflows list
```

## Contributing

We welcome contributions! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) guide for details on how to set up the development environment, project structure, testing, and the pull request process.
