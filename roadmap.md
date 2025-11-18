# Claudex Development Notes

## Overview
This document tracks feature ideas, architectural decisions, and implementation notes for the Claudex framework.

## Architecture Considerations

### Session Directory Execution
Execute `claude` directly in the session directory to give it focused access to accumulated context. Benefits:
- Direct access to all session-specific context
- Avoids overloading with irrelevant project files
- Subagents inherit the same focused context
- Reduces context pollution from the broader project

## Feature Roadmap

### Context Management
- [ ] **`/reload-context` command**: Refresh session context without losing state
  - Clear current context (delete Transcript Path or use built-in `/clear`)
  - Execute command.md to reload context files from session folder
  - Restore the selected profile via post-command hook
  - Note: Session management with `/exit` handling and resume is already functional

- [ ] **Reload session state**: Load additional documentation added to session folder
  - Implement as custom `/command`
  - Useful when documentation is updated mid-session

### Session Lifecycle Hooks
- [ ] **`/exit` hook**: Capture session end
  - Summarize session and create resumption file
  - Run in background to avoid blocking user
  - Enables heavy processing without UX impact

- [ ] **SubagentStop hook**: Capture agent execution results
  - Update session context when `message.stop_reason == end_turn`
  - **Decision needed**: Should all end_prompts update documentation, or only agent executions?
  - **Smart documentation**: Only create/update docs when truly valuable
    - Criteria TBD: significance of changes, new information volume, relevance to session goals, etc.
    - Avoid documentation bloat from minor changes

### Session Management
- [x] **Session and agent selection** at startup
- [ ] **Session prompt flow**:
  - Arrow-key navigation
  - Option 1: Create new session (enter description → generate name → create folder)
  - Option 2: Resume existing session
  - Option 3: Ephemeral session (no folder, no persistence)

### Agent System
- [ ] **Generic agents**: Base agents that can be extended with custom documentation
- [ ] **Multi-tool support**: Execute prompts with alternative AI tools (Codex, Gemini, etc.)
  - Custom commands like `/gemini`, `/codex`
  - Build optimized prompts via hooks
  - Return results to Claude for integration

### Hook System
- [ ] **Implement comprehensive hooks**: All hooks receive `session_id` as input
  - **preCompact hook**: Intercept before context compaction
    - Update session data
    - Abort compaction
    - Restart agent with fresh session data
  - **Command hooks**: Pass session path to all commands
    - Ensures agents always reference correct context
    - Useful for spawning subagents with full context
    - Example prompt: "You are working on this workspace: /path/to/session"
  - **Custom command hooks**: Integrate external tools
    - Build prompts and execute alternative AI tools
    - Get second opinions on complex problems
    - Feed results back to Claude

### Infrastructure
- [ ] **Enable MCPs** during framework installation
- [ ] **Documentation injection**: Design user-friendly method to add custom docs to sessions

## Installation Requirements

The installation script should:
- Create symbolic links from `.claudex` to `.claude` folders
- Create symbolic links from `.claudex` to `.cursor` folders
- Create symbolic links from project path to `.claude`, `.cursor`, and other required folders
- Copy and install the `claudex` script to `$PATH`

## Directory Structure

```
.claudex/
├── agents/       # Agent definitions
├── tasks/        # Task templates
├── templates/    # Session templates
└── sessions/     # Active and archived sessions
```

## Technical Notes

### Session Management Commands
```bash
# Create and activate a session
session_id=$(claude --system-prompt "prompt" "activate" --output-format json | jq -r '.session_id')

# Resume a session
claude --resume $session_id
```

## Current Issues & Feedback

### Issues
- **File organization**: Claude currently generates documents anywhere; should be constrained to session folder

### What's Working Well
- ✅ Team Lead agent effectively delegates to Architect Assistant and Engineer