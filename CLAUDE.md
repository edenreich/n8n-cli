## Instructions

This project is a CLI tool for synchronizing workflows between local JSON files and an n8n instance. It allows you to export workflows from n8n to local files and import them back, ensuring that your workflow configurations are consistent across different environments.

## Project Structure

- `cmd/`: Contains all command implementations using Cobra
- `hack/workflows/`: Sample workflow JSON files for testing
- `Taskfile.yaml`: Task definitions for development
- `main.go`: Entry point for the CLI application

## Tools

You have the following tools available for development:

- [Taskfile](https://taskfile.dev/#/): A task runner for automating development tasks (preferred over Makefile).
- [Go](https://golang.org/): The programming language used for this project.
- [Cobra](https://github.com/spf13/cobra): CLI framework used for command structure.
- [Cobra CLI](https://github.com/spf13/cobra-cli): Code generation tool for new commands.
- [Context7](https://github.com/upstash/context7): An MCP server for documentation fetching.

## Development Workflow

1. Start by writing unit tests for your new features or commands (Using Test-driven development).
2. Always review `Taskfile.yaml` for available development tasks.
3. Always use existing types from n8n for consistency.
4. Always prefer using early returns in your code to avoid deep nesting.
5. Use `task cli -- <args>` to run the CLI during development.
6. Always run `task lint` to verify and fix code errors before proceeding with the build.
7. Use `task build` to build the CLI with proper version information.
8. Use Cobra-CLI to generate new commands: `cobra-cli add <command-name>`
9. Always fetch the latest documentation using context7 and fall back to fetch.
10. When adding new commands, make sure to properly handle flags and implement the Run function.
