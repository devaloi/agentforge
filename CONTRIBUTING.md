# Contributing to agentforge

Thank you for your interest in contributing! This document covers how to get started.

## Development Setup

```bash
git clone https://github.com/devaloi/agentforge.git
cd agentforge
go build ./...
go test ./...
```

### Prerequisites

- Go 1.22+
- golangci-lint (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`)

## Running Tests

```bash
make test        # Run all tests (no real LLM calls — uses mock provider)
make lint        # Run linter
make all         # Format, vet, lint, test, build
```

All tests use a mock LLM provider, so no API keys are needed for testing.

## Adding a New Agent Type

1. Add the agent definition to `config/agents.yaml`
2. No Go code changes needed — agents are configuration-driven

## Adding a New Tool

1. Implement the `Tool` interface in `internal/tools/`
2. Register it in the tool registry
3. Add a schema entry to `config/tools.yaml`
4. Write tests in `internal/tools/`

## Pull Request Guidelines

- Keep PRs focused — one feature or fix per PR
- Add tests for new functionality
- Run `make all` before submitting
- Update README if adding a new feature

## Reporting Issues

Open an issue on GitHub with:
- Go version (`go version`)
- Steps to reproduce
- Expected vs actual behavior
