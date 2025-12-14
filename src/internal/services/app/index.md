# App Package

Main application container for claudex CLI.

## Core

- `app.go` - App struct with Init/Run/Close lifecycle, config loading, logging setup, and hook prompt
- `deps.go` - Dependencies struct for dependency injection (FS, Cmd, Clock, UUID, Env)

## Launch

- `launch.go` - Session launch modes (new, resume, fork, fresh, ephemeral) and Claude CLI invocation
- `session.go` - Session selector TUI and handlers for new/resume/fork workflows

## Tests

- `app_test.go` - Tests for App initialization and run logic
- `launch_test.go` - Tests for launch modes and Claude invocation
nd run logic
- **launch_test.go** - Tests for session launching behavior
