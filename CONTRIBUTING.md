# Contributing to Claudex

Thank you for your interest in contributing to Claudex! This document provides guidelines for contributing to the project.

## Reporting Bugs

When reporting bugs, please include:
- A clear description of the issue
- Steps to reproduce the problem
- Expected vs actual behavior
- Your environment (OS, Go version, Claudex version)
- Relevant logs or error messages

Open an issue on GitHub with the `bug` label.

## Suggesting Features

Feature suggestions are welcome! Please:
- Check existing issues to avoid duplicates
- Describe the feature and its use case clearly
- Explain why this feature would benefit Claudex users
- Open an issue with the `enhancement` label

## Submitting Pull Requests

1. **Fork and Clone**: Fork the repository and clone it locally
2. **Create a Branch**: Use a descriptive branch name (e.g., `fix-config-parsing`, `add-session-export`)
3. **Make Changes**: Follow the code style guidelines below
4. **Write Tests**: Add or update tests for your changes
5. **Run Tests**: Ensure all tests pass with `go test ./...`
6. **Format Code**: Run `go fmt ./...` and `go vet ./...`
7. **Commit**: Write clear, concise commit messages
8. **Push**: Push to your fork and submit a pull request
9. **Describe**: Explain what your PR does and why

## Code Style Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go) conventions
- Use `gofmt` for formatting (non-negotiable)
- Write clear, self-documenting code
- Add godoc comments for exported functions and types
- Keep functions focused and reasonably sized
- Handle errors explicitly - never ignore them
- Use meaningful variable and function names

## Testing Requirements

- Write table-driven tests for new functionality
- Maintain or improve code coverage
- Test edge cases and error conditions
- Use `t.Run()` for subtests
- Ensure tests are deterministic and don't rely on timing

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/claudex.git
cd claudex

# Install dependencies
go mod download

# Run tests
go test ./...

# Build the binary
go build -o claudex ./cmd/claudex
```

## Questions?

If you have questions about contributing, feel free to open an issue for discussion.

Thank you for contributing to Claudex!
