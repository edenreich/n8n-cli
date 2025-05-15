# N8N-CLI TODO List

This document outlines essential features for the n8n-cli tool based on the generated OpenAPI types. These features cover the core functionality needed to effectively manage n8n workflows between local development and n8n instances.

## Core Features

### Workflow Management

- [x] Import workflows from n8n instance to local files
- [x] Synchronize local workflow files to n8n instance
- [x] Activate/deactivate workflows
- [x] Delete workflows from n8n instance
- [ ] Get execution history for workflows
- [ ] Implement validate command to apply static analysis on workflow files - should help to identify issues before syncing to n8n instance
- [] List all workflows with filter capabilities (by name, tags, active status)

### Credentials Management

- [ ] List credentials from n8n instance
- [ ] Apply credentials from Github to n8n instance - it seems that only creation is possible from openapi, which makes sense for security reasons. Have to check how to get the credentials reference correctly so it could be used in the workflow files

### Workflow Execution

- [ ] Execute a workflow manually
- [ ] Retrieve execution results
- [ ] Monitor execution status

### Variables Management

- [ ] List variables
- [ ] Export variables to local files
- [ ] Import variables from local files
- [ ] Set/update variable values

### Tags Management

- [ ] List tags
- [ ] Add/remove tags to workflows
- [ ] Create new tags

### Project Management

- [ ] List projects
- [ ] Create new projects
- [ ] Transfer workflows between projects

### Audit & Security

- [ ] Generate audit reports for workflows
- [ ] Validate workflow files locally before upload

### Configuration & Setup

- [ ] Initialize local configuration
- [ ] Set/update n8n instance URL
- [ ] Set/update API key
- [ ] Configure default project
- [ ] Enable verbose logging

## Technical Enhancements

- [ ] Add validation for local workflow files
- [ ] Implement retry logic for API requests
- [ ] Add support for multiple n8n instances (profiles)
- [ ] Create workspace configuration for team collaboration
- [ ] Add support for environment-specific variables
- [x] Add check-dirty feature to CI pipeline to detect uncommitted generated files

## Documentation

- [ ] Generate command reference
- [ ] Create examples for common workflows
- [ ] Document best practices for workflow version control
