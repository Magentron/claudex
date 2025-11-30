# Contributing to Claudex

Thank you for your interest in contributing to Claudex! This document provides guidelines for contributing.

## Development Setup

1. **Prerequisites**
   - Go 1.21 or higher
   - Claude CLI installed
   - Git

2. **Clone and build**
   ```bash
   git clone https://github.com/YOUR_USERNAME/claudex.git
   cd claudex/claudex
   make build
   ```

3. **Run locally**
   ```bash
   make run
   ```

## Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Keep functions focused and small
- Write descriptive commit messages
- Add comments for non-obvious logic

## Pull Request Process

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** with clear, focused commits
3. **Test your changes** thoroughly
4. **Update documentation** if needed
5. **Submit a pull request** with a clear description

### PR Guidelines

- Keep PRs focused on a single change
- Include relevant issue numbers in the description
- Ensure the build passes
- Respond to review feedback promptly

## Reporting Issues

When reporting bugs, please include:

- Claudex version (`claudex --version` if available, or commit hash)
- Operating system and version
- Steps to reproduce the issue
- Expected vs actual behavior
- Any error messages or logs

## Feature Requests

Feature requests are welcome! Please:

- Check existing issues first to avoid duplicates
- Describe the use case and expected behavior
- Explain why this feature would be valuable

## Agent Profiles

When contributing new agent profiles:

1. Place role templates in `profiles/roles/`
2. Place skill files in `profiles/skills/`
3. Pre-built agents go in `profiles/agents/`
4. Follow the existing frontmatter format (name, description, model, color)
5. Test the profile with actual Claude sessions

## Questions?

Feel free to open an issue for any questions about contributing.
