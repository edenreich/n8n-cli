# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a CLI tool for synchronizing workflows between local JSON files and an n8n instance. It allows you to export workflows from n8n to local files and import them back, ensuring that your workflow configurations are consistent across different environments.

## Architecture

### Core Components

- **`cmd/`**: Contains all command implementations using Cobra CLI framework
  - `root.go`: Main CLI entry point with global flags and configuration
  - `workflows.go`: Parent command for workflow operations
  - `workflows/`: Subcommands for workflow management (list, sync, refresh, activate, deactivate)
- **`n8n/`**: n8n API client and data handling
  - `client.go`: HTTP client for n8n API interactions
  - `generated_types.go`: Auto-generated types from n8n OpenAPI specification
  - `encoder.go`: JSON/YAML encoding utilities for workflow files
  - `interface.go`: Client interface for dependency injection and testing
- **`config/`**: Configuration management using Viper
  - `config.go`: Main configuration initialization
  - `dotenv.go`: Environment variable loading from .env files
  - `version.go`: Build-time version information
- **`logger/`**: Centralized logging using Zap
- **`tests/`**: Comprehensive test suite
  - `unit/`: Unit tests with mocked dependencies
  - `integration/`: Integration tests against real n8n instances

### Key Patterns

- **Dependency Injection**: Uses interfaces (`n8n.Client`) with fake implementations for testing
- **Configuration**: Viper handles flags, environment variables, and .env files with precedence
- **Error Handling**: Early returns to avoid nested conditionals
- **Test-Driven Development**: Comprehensive unit and integration test coverage
- **Code Generation**: Uses counterfeiter for generating test fakes and oapi-codegen for API types

## Development Commands

### Essential Commands

```bash
# Run CLI during development
task cli -- <args>

# Build with version info
task build

# Run tests
task test-unit          # Unit tests only
task test-integration   # Integration tests only  
task test              # All tests

# Code quality
task lint              # Run golangci-lint
task generate          # Generate mocks and interfaces
```

### OpenAPI Integration

```bash
# Download latest n8n OpenAPI spec
task oas-download

# Generate Go types from OpenAPI spec
task oas-generate

# Lint OpenAPI specification
task oas-lint
```

## Development Workflow

1. **Start with tests**: Use Test-Driven Development - write unit tests first
2. **Check available tasks**: Always review `Taskfile.yaml` for available development tasks
3. **Use existing types**: Always use existing types from n8n for consistency (`n8n/generated_types.go`)
4. **Early returns**: Prefer early returns in code to avoid deep nesting
5. **Run quality checks**: Always run `task lint` before building
6. **Use Cobra CLI**: Generate new commands with `cobra-cli add <command-name>`
7. **Documentation**: Check Context7 documentation and fallback to fetch for n8n API details

## Important Files

- **`Taskfile.yaml`**: All development tasks and build configuration
- **`openapi.yml`**: n8n API specification (auto-downloaded)
- **`go.mod`**: Dependencies including Cobra, Viper, Zap, and testing frameworks
- **`.github/copilot-instructions.md`**: Additional development guidelines
- **`CHANGELOG.md`**: Auto-generated from commit messages (never edit directly)

## Testing Strategy

- **Mocking**: Uses counterfeiter to generate fakes for the n8n.Client interface
- **Integration Tests**: Test against real n8n instances with proper setup/teardown
- **Coverage**: Separate coverage reports for unit and integration tests
- **Test Helpers**: Shared utilities in `tests/unit/helpers.go` and `tests/integration/helpers.go`

## Configuration Management

The CLI supports multiple configuration sources with this precedence:
1. Command-line flags
2. Environment variables
3. `.env` file in current directory

Key environment variables:
- `N8N_API_KEY`: n8n API authentication token
- `N8N_INSTANCE_URL`: n8n instance URL
- `DEBUG`: Enable debug logging

## Build System

Uses Taskfile with dynamic version information:
- Version: Git tag or "dev"
- Commit: Short git hash
- Build Date: UTC timestamp

Version information is injected at build time using Go's `-ldflags`.