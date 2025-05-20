# n8n Workflow Repository

This repository serves as a version-controlled storage and synchronization system for [n8n](https://n8n.io/) workflows. It enables GitOps practices for n8n workflow management, allowing for code reviews, version history, and automated synchronization between GitHub and n8n instances.

## 📋 Table of Contents

- [Features](#-features)
- [Repository Structure](#-repository-structure)
- [Synchronization Workflows](#-synchronization-workflows)
  - [Sync Workflow (n8n → GitHub)](#sync-workflow-n8n--github)
  - [Restore Workflow (GitHub → n8n)](#restore-workflow-github--n8n)
- [Setup Requirements](#️-setup-requirements)
- [Development Tools](#-development-tools)
- [Development Process](#-development-process)

## 🚀 Features

- **Version-controlled workflows**: Store n8n workflows as YAML files in Git
- **Bidirectional synchronization**: Automated sync between GitHub and n8n cloud instance
- **GitHub Actions integration**: Automate workflow deployment and updates
- **Pull Request workflow**: Changes from cloud instance to GitHub go through PR review
- **Easy restoration**: Quickly restore workflows to n8n from GitHub repository

## 📂 Repository Structure

```
.
├── .github/workflows/    # GitHub Actions workflow definitions
│   ├── sync.yml         # Syncs workflows from n8n instance to GitHub (creates PRs)
│   └── restore.yml      # Restores workflows from GitHub to n8n instance
├── workflows/           # YAML definitions of n8n workflows
│   └── *.yaml           # Individual workflow definitions
└── README.md            # This file
```

## 🔄 Synchronization Workflows

This repository implements two GitHub Actions workflows for managing n8n workflows:

### Sync Workflow (n8n → GitHub)

The \`sync.yml\` workflow automatically pulls the latest workflow definitions from your n8n instance and creates a Pull Request with those changes. This allows for:

- Code review of workflow changes made in the n8n UI
- Version history of all workflow modifications
- Safe integration of changes through the PR process

Triggered:

- Automatically on push to main branch (for relevant paths)
- Manually via GitHub Actions workflow_dispatch

### Restore Workflow (GitHub → n8n)

The \`restore.yml\` workflow pushes workflow definitions from the GitHub repository to your n8n instance. This allows for:

- Disaster recovery
- Deploying workflows to new instances
- Rolling back to previous versions

Triggered:

- Manually via GitHub Actions workflow_dispatch

## 🛠️ Setup Requirements

To use this repository, you need:

1. **n8n Instance**: A running n8n instance (cloud or self-hosted)
2. **API Key**: An n8n API key with appropriate permissions
3. **GitHub Secrets**:
   - \`N8N_API_KEY\`: Your n8n API key
   - \`N8N_INSTANCE_URL\`: The URL of your n8n instance
   - \`GROQ_API_KEY\`: Your GROQ API key for Pull Request changes summary

## 🧰 Development Tools

This repository uses:

- [n8n CLI](https://github.com/edenreich/n8n-cli) for workflow management
- GitHub Actions for automation
- [create-pull-request](https://github.com/peter-evans/create-pull-request) GitHub action for PR creation

## 🤝 Development Process

To contribute:

1. Make changes to workflows in your n8n instance
2. Run the sync workflow to create a PR with those changes
3. Review the PR and merge it to maintain version history

Alternatively, you can make changes directly to the YAML files and use the restore workflow to push them to your n8n instance.
