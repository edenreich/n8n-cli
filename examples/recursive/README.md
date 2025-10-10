# Recursive Directory Workflows Example

This example demonstrates how to use the `--recursive` flag with n8n-cli to manage workflows organized in a nested directory structure. This is particularly useful for organizing workflows by feature, team, or project in subdirectories.

## Table of Contents

- [Overview](#overview)
- [Directory Structure](#directory-structure)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Usage Examples](#usage-examples)
  - [Sync Workflows Recursively](#sync-workflows-recursively)
  - [Refresh Workflows Recursively](#refresh-workflows-recursively)
  - [Non-Recursive vs Recursive](#non-recursive-vs-recursive)
- [Example Workflows](#example-workflows)
- [Cleanup](#cleanup)

## Overview

The `--recursive` flag enables n8n-cli to scan subdirectories when syncing or refreshing workflows. This allows you to:

- **Organize workflows by feature or team** in separate subdirectories
- **Maintain logical groupings** without flattening your directory structure
- **Preserve directory hierarchy** when refreshing workflows from n8n
- **Scale workflow management** for large projects with many workflows

## Directory Structure

```
examples/recursive/
├── docker-compose.yaml          # n8n instance setup
├── README.md                    # This file
└── workflows/                   # Root workflows directory
    ├── feature_A/               # Feature A workflows
    │   ├── contact_form_processor.json
    │   └── ai_response_generator.json
    └── feature_B/               # Feature B workflows
        ├── data_sync.json
        └── sub_feature_C/       # Nested subdirectory
            └── error_monitoring.json
```

## Prerequisites

- Docker and Docker Compose installed
- Go 1.25+ installed
- Git (to clone this repository)

## Quick Start

1. **Build the n8n-cli binary:**

   From the repository root, run:

   ```bash
   go build -o examples/recursive/bin/n8n .
   ```

2. **Start the n8n instance:**

   ```bash
   docker compose up -d
   ```

3. **Wait for n8n to be ready:**

   ```bash
   # Check health status
   docker compose ps

   # Or watch logs
   docker compose logs -f n8n
   ```

4. **Access n8n:**

   Open http://localhost:5678 in your browser and complete the initial setup.

5. **Set your n8n credentials:**

   ```bash
   export N8N_API_KEY="your-api-key-here"
   export N8N_INSTANCE_URL="http://localhost:5678"
   ```

   Or create a `.env` file:

   ```bash
   cat > .env <<EOF
   N8N_API_KEY=your-api-key-here
   N8N_INSTANCE_URL=http://localhost:5678
   EOF
   ```

## Usage Examples

### Sync Workflows Recursively

Sync all workflow files from the `workflows/` directory and its subdirectories to your n8n instance:

```bash
./bin/n8n workflows sync --directory workflows --recursive
```

**What happens:**
- Scans `workflows/` and all subdirectories
- Creates or updates workflows in n8n
- Activates workflows marked as `"active": true`
- Preserves the workflow IDs from the JSON files

**Without the `--recursive` flag:**
```bash
# This would only sync files directly in workflows/, not in subdirectories
./bin/n8n workflows sync --directory workflows
```

### Refresh Workflows Recursively

Pull the latest workflow state from n8n and update local files while preserving directory structure:

```bash
# Refresh all workflows, maintaining directory structure
./bin/n8n workflows refresh --directory workflows --recursive --all

# Refresh only workflows that exist in local directories
./bin/n8n workflows refresh --directory workflows --recursive
```

**What happens with `--recursive`:**
- Updates existing workflow files in their current subdirectories
- When using `--all`, creates new workflow files in the root directory
- Preserves the directory structure of existing workflows

**With additional options:**
```bash
# Overwrite local changes with remote state
./bin/n8n workflows refresh --directory workflows --recursive --overwrite

# Output in YAML format instead of JSON
./bin/n8n workflows refresh --directory workflows --recursive --output yaml
```

### Non-Recursive vs Recursive

**Non-recursive mode** (default):
```bash
./bin/n8n workflows sync --directory workflows
```
- Only processes files directly in `workflows/`
- Ignores `feature_A/`, `feature_B/`, and `sub_feature_C/`
- Result: 0 workflows synced

**Recursive mode:**
```bash
./bin/n8n workflows sync --directory workflows --recursive
```
- Processes all files in `workflows/` and subdirectories
- Includes `feature_A/`, `feature_B/`, and `sub_feature_C/`
- Result: 4 workflows synced

### Dry Run

Preview what would happen without making changes:

```bash
./bin/n8n workflows sync --directory workflows --recursive --dry-run
```

### Sync with Refresh

Sync local workflows to n8n, then refresh local state:

```bash
./bin/n8n workflows sync --directory workflows --recursive --refresh --all
```

## Example Workflows

This example includes 4 sample workflows demonstrating different use cases:

### Feature A: Contact Form Processing

1. **Contact Form Processor** (`feature_A/contact_form_processor.json`)
   - Receives form submissions via webhook
   - Processes and validates form data
   - Sends email notifications
   - Status: Active

2. **AI Response Generator** (`feature_A/ai_response_generator.json`)
   - Chat webhook endpoint
   - Generates AI responses using OpenAI
   - Returns responses to webhook caller
   - Status: Active

### Feature B: Data Operations

3. **Data Sync Workflow** (`feature_B/data_sync.json`)
   - Runs every 6 hours
   - Fetches data from external API
   - Stores in PostgreSQL database
   - Status: Inactive

4. **Error Monitoring & Alerts** (`feature_B/sub_feature_C/error_monitoring.json`)
   - Receives error reports via webhook
   - Evaluates error severity
   - Sends Slack alerts for critical errors
   - Status: Active

## Cleanup

When you're done with the example:

```bash
# Stop and remove the n8n container
docker compose down

# Remove all data (optional)
docker compose down -v
```

---

**Note:** This example uses a local n8n instance for demonstration. In production, you would typically point to a hosted n8n instance using your actual `N8N_INSTANCE_URL` and `N8N_API_KEY`.
