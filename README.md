# n8n-cli

Command line interface for managing n8n workflows.

## Installation

```bash
go install github.com/edenreich/n8n-cli@latest
```

## Configuration

Create a `.env` file with the following variables:

```
N8N_API_KEY=your_n8n_api_key
N8N_INSTANCE_URL=https://your-instance.n8n.cloud
```

You can generate an API key in the n8n UI under Settings > API.

## Commands

### Sync

Synchronize JSON workflows from a local directory to an n8n instance:

```bash
n8n-cli sync --directory hack/workflows
```

Options:

- `--directory, -d`: Directory containing workflow JSON files (default: "hack/workflows")
- `--activate-all, -a`: Activate all workflows after synchronization
- `--dry-run, -n`: Show what would be done without making changes
- `--verbose, -v`: Show detailed output during synchronization

Example:

```bash
# Sync all workflows and activate them
n8n-cli sync --activate-all

# Test without making changes
n8n-cli sync --dry-run
```

### Import

Import workflows from an n8n instance to local JSON files:

```bash
n8n-cli import --directory hack/workflows
```

Options:

- `--directory, -d`: Directory to save workflow JSON files (default: "hack/workflows")
- `--workflow-id, -w`: ID of a specific workflow to import
- `--all, -a`: Import all workflows (default if no workflow-id is specified)
- `--dry-run, -n`: Show what would be done without making changes
- `--verbose, -v`: Show detailed output during import

Example:

```bash
# Import all workflows from n8n
n8n-cli import

# Import a specific workflow by ID
n8n-cli import --workflow-id 123

# Test without making changes
n8n-cli import --dry-run
```

## Workflow File Structure

Workflow files should be valid n8n workflow JSON files. The sync command will:

1. Create new workflows for files without an ID or with an ID that doesn't exist on the n8n instance
2. Update existing workflows that have a matching ID
3. Activate workflows based on the "active" property in the JSON file or if --activate-all is used
