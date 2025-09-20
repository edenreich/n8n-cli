# Contributing to n8n-cli

Thank you for considering contributing to n8n-cli! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
  - [Prerequisites](#prerequisites)
  - [Getting Started](#getting-started)
  - [Development Environment](#development-environment)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Running Tests](#running-tests)
- [Adding New Commands](#adding-new-commands)
- [Code Style and Guidelines](#code-style-and-guidelines)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Development Environment Setup

### Prerequisites

- Go 1.25 or higher
- [Task](https://taskfile.dev/) - Task runner for development automation
- [Cobra CLI](https://github.com/spf13/cobra-cli) - For generating new commands
- Git

### Getting Started

1. Fork the repository on GitHub
2. Clone your fork to your local machine:

   ```bash
   git clone https://github.com/YOUR-USERNAME/n8n-cli.git
   cd n8n-cli
   ```

3. Add the original repository as an upstream remote:

   ```bash
   git remote add upstream https://github.com/edenreich/n8n-cli.git
   ```

4. Install development dependencies:
   ```bash
   task testing-deps
   ```

### Development Environment

We strongly recommend using the provided dev container environment for development, as it comes pre-configured with all the necessary tools and dependencies for this project. The dev container includes Go, Task, Cobra CLI, and other relevant tools.

If you are using Visual Studio Code, you can easily start the dev container by:

1. Installing the "Remote - Containers" extension
2. Opening the project folder
3. Clicking on the green button in the lower-left corner and selecting "Reopen in Container"

For more information on dev containers, see the [VS Code Dev Containers documentation](https://code.visualstudio.com/docs/devcontainers/containers).

## Project Structure

The project follows a standard Go module structure:

- `cmd/`: Contains all command implementations using Cobra
  - `cmd/workflows/`: Subcommands for managing workflows
- `config/`: Configuration handling
- `hack/`: Example workflow files and development utilities
- `n8n/`: n8n API client and types
- `tests/`: Test files
  - `tests/unit/`: Unit tests
  - `tests/integration/`: Integration tests

## Development Workflow

1. Create a new branch for your feature or bugfix:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes, following the Test-Driven Development approach:

   - Write tests first in the appropriate test directory
   - Implement the feature to make the tests pass
   - Refactor the code if needed

3. Run the linter to check for code quality issues:

   ```bash
   task lint
   ```

4. Run tests to ensure your changes work correctly:

   ```bash
   task test
   ```

5. During development, you can run the CLI with:
   ```bash
   task cli -- <args>
   ```
   For example:
   ```bash
   task cli -- workflows list
   ```

## Running Tests

The project includes both unit and integration tests:

- Run unit tests only:

  ```bash
  task test-unit
  ```

- Run integration tests only:

  ```bash
  task test-integration
  ```

- Run all tests:
  ```bash
  task test
  ```

## Adding New Commands

To add a new command to the CLI:

1. Generate a command scaffold using Cobra CLI:

   ```bash
   cobra-cli add <command-name>
   ```

2. For subcommands:

   ```bash
   cobra-cli add <subcommand-name> -p <parent-command>Cmd
   ```

3. Implement the command logic in the generated file's `Run` function.

4. Create unit and integration tests in the respective test folders.

## Code Style and Guidelines

- Follow Go's official [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for style guidance
- Use meaningful variable and function names that reflect their purpose
- Write comments for exported functions, types, and packages
- Ensure proper error handling with meaningful error messages
- Keep functions focused and concise, following the Single Responsibility Principle

## Pull Request Process

1. Update your branch with the latest changes from upstream:

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Ensure your code passes all tests and linting:

   ```bash
   task lint
   task test
   ```

3. Push your changes to your fork:

   ```bash
   git push origin feature/your-feature-name
   ```

4. Create a Pull Request through GitHub's interface with:

   - A clear title and description of the changes
   - References to any related issues
   - Explanation of how your changes have been tested

5. Address any feedback or requested changes from the code review.

## Release Process

This project uses [semantic-release](https://github.com/semantic-release/semantic-release) for automated version management, release creation, and release notes generation. The release process is fully automated based on conventional commit messages when code is merged to the main branch.

---

Thank you for contributing to n8n-cli! Your efforts help improve the tool for everyone.
