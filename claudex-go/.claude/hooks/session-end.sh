#!/bin/bash

# SessionEnd hook for debugging/testing
# This hook is triggered when a Claude Code session ends

LOG_FILE="/Users/maikel/Workspace/Pelago/bmad/.claude/hooks/session-end.log"

# Prevent infinite loop - check if we're already in a hook
if [ "$IN_SESSION_END_HOOK" = "1" ]; then
    echo "Skipping hook (already in hook context)" >> "$LOG_FILE"
    exit 0
fi

# Set flag to prevent recursive hook calls
export IN_SESSION_END_HOOK=1

# Read JSON input from stdin
INPUT_JSON=$(cat)

# Extract session_id using basic text processing (jq-free approach)
SESSION_ID=$(echo "$INPUT_JSON" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
REASON=$(echo "$INPUT_JSON" | grep -o '"reason":"[^"]*"' | cut -d'"' -f4)

# Run the slow operations in a completely detached background process
# Using nohup and redirecting stdin from /dev/null to fully detach
nohup bash -c '
    LOG_FILE="/Users/maikel/Workspace/Pelago/bmad/.claude/hooks/session-end.log"
    SESSION_ID="'"$SESSION_ID"'"
    REASON="'"$REASON"'"

    # Create log entry with timestamp
    echo "===========================================================" >> "$LOG_FILE"
    echo "SessionEnd hook triggered at: $(date "+%Y-%m-%d %H:%M:%S")" >> "$LOG_FILE"
    echo "Session ID: $SESSION_ID" >> "$LOG_FILE"
    echo "Reason: $REASON" >> "$LOG_FILE"
    echo "Current directory: $(pwd)" >> "$LOG_FILE"
    echo "User: $USER" >> "$LOG_FILE"
    echo "Session ended successfully" >> "$LOG_FILE"
    echo "" >> "$LOG_FILE"

    # Generate a random sentence using Claude (using full path)
    echo "Generating random sentence from Claude..." >> "$LOG_FILE"
    export IN_SESSION_END_HOOK=1
    RANDOM_SENTENCE=$(/opt/homebrew/bin/claude -p "Generate a single random creative sentence. Only output the sentence, nothing else." 2>&1)

    # Add the random sentence to the log
    echo "Random sentence: $RANDOM_SENTENCE" >> "$LOG_FILE"
    echo "===========================================================" >> "$LOG_FILE"
    echo "" >> "$LOG_FILE"
' </dev/null >/dev/null 2>&1 &

# Disown the background process so it's not tied to this shell
disown

# Exit immediately
exit 0
