#!/bin/bash

# SubagentStop hook for debugging/testing
# This hook is triggered when a subagent (Task tool) completes

LOG_FILE="/Users/maikel/Workspace/Pelago/bmad/.claude/hooks/subagent-stop.log"

# Read JSON input from stdin
INPUT_JSON=$(cat)

# Extract session_id using basic text processing (jq-free approach)
SESSION_ID=$(echo "$INPUT_JSON" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
TRANSCRIPT_PATH=$(echo "$INPUT_JSON" | grep -o '"transcript_path":"[^"]*"' | cut -d'"' -f4)

# Create log entry with timestamp
echo "===========================================================" >> "$LOG_FILE"
echo "SubagentStop hook triggered at: $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
echo "Session ID: $SESSION_ID" >> "$LOG_FILE"
echo "Transcript Path: $TRANSCRIPT_PATH" >> "$LOG_FILE"
echo "Current directory: $(pwd)" >> "$LOG_FILE"
echo "User: $USER" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"
echo "Full JSON input:" >> "$LOG_FILE"
echo "$INPUT_JSON" >> "$LOG_FILE"
echo "===========================================================" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# Exit successfully
exit 0
