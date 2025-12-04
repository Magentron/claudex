# Claudex Session Manager

A modern, interactive session manager for Claude Code, built in Go with a beautiful TUI.

## Features

- 🎨 **Beautiful TUI** - Modern interface built with Bubble Tea framework
- 📋 **Session Management** - Create, resume, and manage work sessions
- 🤖 **AI-Generated Names** - Claude suggests session names based on descriptions
- ⚡ **Ephemeral Mode** - Work without saving session data
- 🎭 **Profile Selection** - Choose different Claude configurations
- 🔍 **Fuzzy Search** - Filter sessions and profiles by typing `/`
- 📊 **Session Metadata** - Track descriptions and creation dates
- 📱 **Responsive** - Adapts to terminal width automatically
- ⌨️ **Keyboard Controls** - Full keyboard navigation (↑↓, Enter, q, Ctrl+C)

## Installation

```bash
cd claudex
make install
```

Or manually:
```bash
go mod tidy
go build -o claudex-session main.go
chmod +x claudex-session
```

## Usage

```bash
./claudex-session
```

The program will guide you through:
1. **Session Selection** - Create new, use ephemeral mode, or resume an existing session
2. **Profile Selection** - Choose a Claude profile/configuration
3. **Launch Claude** - Automatically launches Claude Code with your selections

### Keyboard Controls

**Session/Profile Lists:**
- `↑/↓` - Navigate menu
- `Enter` - Select option
- `/` - Start fuzzy search
- `q` or `Ctrl+C` - Quit

**Create New Session:**
- Type description and press `Enter`
- Claude will generate a slug-based name automatically

## Documentation Loading with --doc Flag

The `--doc` flag allows you to provide context documentation to Claude. For optimal performance, use the **index file pattern** instead of passing entire directories.

### The Problem with Directories

```bash
# Not recommended: Loads ALL files immediately
claudex --doc=./docs/
```

When you pass a directory, Claude loads **all documentation files** into the context window upfront. This:
- Wastes context tokens on irrelevant docs
- Slows down agent initialization
- Makes it harder for Claude to find relevant information

### The Solution: Index File Pattern

Create an index file (e.g., `docs/INDEX.md`) that serves as a table of contents:

```bash
# Recommended: Load index, read details on-demand
claudex --doc=./docs/INDEX.md
```

**Benefits:**
1. Minimal initial context usage (only the index loads)
2. Claude reads detailed specs on-demand when relevant
3. Better navigation - Claude understands doc structure before diving in
4. Faster startup and more efficient token usage

### Example Index File Structure

```markdown
# Project Documentation Index

## Architecture
- [System Overview](./architecture/overview.md) - High-level system design and components
- [Database Schema](./architecture/database.md) - Database tables and relationships
- [API Gateway](./architecture/gateway.md) - Request routing and middleware

## API Reference
- [REST Endpoints](./api/rest.md) - All REST API endpoints and payloads
- [WebSocket Events](./api/websocket.md) - Real-time event specifications
- [Authentication](./api/auth.md) - OAuth2 flows and token management

## Development
- [Setup Guide](./dev/setup.md) - Local development environment setup
- [Testing Strategy](./dev/testing.md) - How to write and run tests
- [Deployment](./dev/deployment.md) - CI/CD pipeline and production deployment

## Code Standards
- [Style Guide](./standards/style.md) - Code formatting and naming conventions
- [Best Practices](./standards/best-practices.md) - Patterns and anti-patterns
```

### Best Practices

1. **Keep descriptions concise** - Brief one-liners help Claude choose which docs to read
2. **Use relative paths** - Makes the index portable
3. **Organize by category** - Group related docs together
4. **Update regularly** - Keep the index in sync with your documentation
5. **Include context** - Add enough detail so Claude knows when to read each file

## Session Data

Sessions are stored in `./sessions/` with:
- `.description` - Your session description
- `.created` - ISO 8601 timestamp

Example:
```
sessions/
├── auth-refactor/
│   ├── .description
│   └── .created
└── api-performance-fix/
    ├── .description
    └── .created
```

## Environment Variables

The program sets these variables for your Claude session:
- `CLAUDEX_SESSION` - Current session name (or "ephemeral")
- `CLAUDEX_SESSION_PATH` - Full path to session directory

## Directory Structure

```
.
├── claudex-session         # The executable
├── sessions/               # Session storage (auto-created)
└── .profiles/              # Claude profile configurations (required)
    ├── default.md
    ├── architect.md
    └── engineer.md
```

## Session Name Generation

When creating a session:
1. Enter a description (e.g., "Working on user authentication")
2. Claude generates a slug (e.g., "user-auth-module")
3. Falls back to auto-generated slug if Claude unavailable
4. Ensures uniqueness by appending numbers if needed

## Building

Use the Makefile for easy building:

```bash
make              # Build claudex-session
make deps         # Install/update dependencies
make install      # Build and mark executable
make clean        # Remove build artifacts
make run          # Build and run
make help         # Show all targets
```

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/bubbles` - TUI components

## Example Workflow

```bash
$ ./claudex-session

# Beautiful TUI appears
# Select "➕ Create New Session"
# Enter: "Refactoring authentication module"
# Auto-generated: "auth-refactor"

# Select profile "🎭 principal-architect.md"
# Claude launches with session context

# Later...
$ ./claudex-session

# Select "📁 auth-refactor"
# Continue where you left off
```

## Troubleshooting

**Issue:** Profiles directory not found
```bash
mkdir .profiles
# Add your profile markdown files here
```

**Issue:** Claude command not found
```bash
# Install Claude CLI first
# See: https://claude.ai/cli
```

**Issue:** Terminal artifacts when launching Claude
- This is fixed with proper terminal restoration
- Wait for "Launching Claude..." message before interaction

## Technical Details

- **Framework:** Bubble Tea (Elm architecture for Go)
- **Alt-screen:** Uses alternate screen buffer for clean UI
- **Terminal restoration:** Properly restores terminal state before launching Claude
- **Message passing:** Uses Bubble Tea's message system for state management
- **Fuzzy search:** Built-in filtering for quick navigation

## License

See main project LICENSE

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
