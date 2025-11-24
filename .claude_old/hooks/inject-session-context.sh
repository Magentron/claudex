#!/bin/bash

# PreToolUse hook to inject session context into agent prompts
# Enforces session-scoped documentation for all agents

set -euo pipefail

# Logging configuration
LOG_FILE="/Users/maikel/Workspace/Pelago/bmad/.claude/hooks/inject-session-context.log"

# Logging function
log_message() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "$timestamp | $1" >> "$LOG_FILE"
}

# Start logging for this invocation
echo "===========================================================" >> "$LOG_FILE"
log_message "Hook triggered"

# Read JSON input from stdin
INPUT_JSON=$(cat)

# Extract tool_name using jq - only process Task tool (agent invocations)
TOOL_NAME=$(echo "$INPUT_JSON" | jq -r '.tool_name // ""')
log_message "Tool: $TOOL_NAME"

if [ "$TOOL_NAME" != "Task" ]; then
    # Pass through for non-Task tools
    log_message "Action: Passed through (not Task tool)"
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow"
  }
}
EOF
    exit 0
fi

# Extract session_id using jq
SESSION_ID=$(echo "$INPUT_JSON" | jq -r '.session_id // ""')

if [ -z "$SESSION_ID" ]; then
    # No session ID, skip injection
    log_message "Session ID: (none)"
    log_message "Action: Passed through (no session ID)"
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow"
  }
}
EOF
    exit 0
fi

log_message "Session ID: $SESSION_ID"

# Find session folder using glob pattern: ./sessions/*-{session_id}
SESSION_PATTERN="./sessions/*-${SESSION_ID}"
SESSION_FOLDERS=($(ls -d $SESSION_PATTERN 2>/dev/null || true))

if [ ${#SESSION_FOLDERS[@]} -eq 0 ]; then
    # No session folder found (ephemeral session), skip injection
    log_message "Session Folder: (not found - ephemeral session)"
    log_message "Action: Passed through (no session folder)"
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow"
  }
}
EOF
    exit 0
fi

# Use first match if multiple folders found (log warning to stderr)
if [ ${#SESSION_FOLDERS[@]} -gt 1 ]; then
    echo "WARNING: Multiple session folders found for session_id $SESSION_ID, using first match" >&2
    log_message "Warning: Multiple session folders found, using first match"
fi

SESSION_FOLDER="${SESSION_FOLDERS[0]}"
SESSION_PATH=$(cd "$SESSION_FOLDER" && pwd)
log_message "Session Folder: $SESSION_PATH"

# List existing files in session folder (excluding hidden files)
FILE_LISTING=$(ls -1 "$SESSION_PATH" 2>/dev/null | grep -v '^\.' || echo "")

if [ -z "$FILE_LISTING" ]; then
    FILES_DISPLAY="- (No files yet - you'll be the first to create documentation!)"
else
    FILES_DISPLAY=$(echo "$FILE_LISTING" | sed 's/^/- /')
fi

# Extract original prompt from tool_input using jq
ORIGINAL_PROMPT=$(echo "$INPUT_JSON" | jq -r '.tool_input.prompt // ""')
DESCRIPTION=$(echo "$INPUT_JSON" | jq -r '.tool_input.description // ""')
SUBAGENT_TYPE=$(echo "$INPUT_JSON" | jq -r '.tool_input.subagent_type // ""')
log_message "Subagent Type: $SUBAGENT_TYPE"

# Create injected context using the template from feature definition
# Note: We need to escape special characters for JSON
INJECTED_CONTEXT=$(cat <<CONTEXT_EOF
## SESSION CONTEXT (CRITICAL)

You are working within an active Claudex session. ALL documentation, plans, and artifacts MUST be created in the session folder.

**Session Folder (Absolute Path)**: \`${SESSION_PATH}\`

### MANDATORY RULES for Documentation:
1. ✅ ALWAYS save documentation to the session folder above
2. ✅ Use absolute paths when creating files (Write/Edit tools)
3. ✅ Before exploring the codebase, check the session folder for existing context
4. ❌ NEVER save documentation to project root or arbitrary locations
5. ❌ NEVER use relative paths for documentation files

### Session Folder Contents:
${FILES_DISPLAY}

### Recommended File Names:
- Research documents: \`research-{topic}.md\`
- Execution plans: \`execution-plan-{feature}.md\`
- Analysis reports: \`analysis-{component}.md\`
- Technical specs: \`technical-spec-{feature}.md\`

---

## ORIGINAL REQUEST

${ORIGINAL_PROMPT}
CONTEXT_EOF
)

# Build the updated input JSON using jq for proper JSON construction
OUTPUT=$(jq -n \
    --arg desc "$DESCRIPTION" \
    --arg type "$SUBAGENT_TYPE" \
    --arg prompt "$INJECTED_CONTEXT" \
    '{
        hookSpecificOutput: {
            hookEventName: "PreToolUse",
            permissionDecision: "allow",
            updatedInput: {
                description: $desc,
                subagent_type: $type,
                prompt: $prompt
            }
        }
    }')

log_message "Action: Injected session context"
log_message "---"
log_message "Original Prompt (first 200 chars):"
log_message "${ORIGINAL_PROMPT:0:200}..."
log_message "---"
log_message "Modified Prompt (first 500 chars):"
log_message "${INJECTED_CONTEXT:0:500}..."
log_message "---"

# Output the modified JSON
echo "$OUTPUT"

exit 0
