#!/bin/bash
# notification-hook.sh - Handle Claude Notification hook events
# Triggered when Claude needs user attention (permission, idle, auth, etc.)
#
# Input (JSON via stdin):
# {
#   "session_id": "abc123",
#   "transcript_path": "/path/to/transcript.jsonl",
#   "cwd": "/current/working/directory",
#   "permission_mode": "default",
#   "hook_event_name": "Notification",
#   "message": "Claude needs your permission to use Bash",
#   "notification_type": "permission_prompt"
# }
#
# Notification types: permission_prompt, idle_prompt, auth_success, elicitation_dialog

# === SETUP ===
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Logging configuration
if [ -z "${CLAUDEX_LOG_FILE:-}" ]; then
    LOG_FILE="$PROJECT_ROOT/.claude/hooks/notification-hook.log"
else
    LOG_FILE="$CLAUDEX_LOG_FILE"
    LOG_DIR=$(dirname "$LOG_FILE")
    mkdir -p "$LOG_DIR" 2>/dev/null || true
fi

log_message() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "$timestamp | [hook_notification] $1" >> "$LOG_FILE"
}

echo "===========================================================" >> "$LOG_FILE"
log_message "Hook triggered (Notification)"

# === RECURSION GUARD ===
if [ "$CLAUDE_HOOK_INTERNAL" == "1" ]; then
    log_message "Recursion detected (CLAUDE_HOOK_INTERNAL=1). Exiting."
    exit 0
fi

# === PARSE INPUT ===
INPUT_JSON=$(cat)

SESSION_ID=$(echo "$INPUT_JSON" | jq -r '.session_id // ""')
MESSAGE=$(echo "$INPUT_JSON" | jq -r '.message // ""')
NOTIFICATION_TYPE=$(echo "$INPUT_JSON" | jq -r '.notification_type // ""')
HOOK_EVENT_NAME=$(echo "$INPUT_JSON" | jq -r '.hook_event_name // ""')
CWD=$(echo "$INPUT_JSON" | jq -r '.cwd // ""')

log_message "Session ID: $SESSION_ID"
log_message "Notification Type: $NOTIFICATION_TYPE"
log_message "Message: $MESSAGE"
log_message "CWD: $CWD"

# === FORMAT SESSION NAME ===
# Extract session name from CLAUDEX_SESSION_PATH
SESSION_NAME=$(basename "$CLAUDEX_SESSION_PATH" 2>/dev/null || echo "unknown")
# Format: remove UUID suffix, replace dashes with spaces, title case
FORMATTED_SESSION=$(echo "$SESSION_NAME" | sed -E 's/-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$//' | tr '-' ' ' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')

log_message "Session Name: $FORMATTED_SESSION"

# === NOTIFICATION TYPE MAPPING ===
# Returns "Title|Sound" for each notification type
get_notification_config() {
    local type="$1"
    case "$type" in
        "permission_prompt")
            echo "Permission Required|Blow"
            ;;
        "idle_prompt")
            echo "Claude Waiting|Submarine"
            ;;
        "auth_success")
            echo "Authentication|Glass"
            ;;
        "elicitation_dialog")
            echo "Input Needed|Ping"
            ;;
        *)
            echo "Claude Notification|Ping"
            ;;
    esac
}

# === SEND NOTIFICATION ===
NOTIFICATION_LIB="$SCRIPT_DIR/lib/notification.sh"
if [ -f "$NOTIFICATION_LIB" ]; then
    source "$NOTIFICATION_LIB"

    # Get configuration for this notification type
    CONFIG=$(get_notification_config "$NOTIFICATION_TYPE")
    TYPE_TITLE=$(echo "$CONFIG" | cut -d'|' -f1)
    SOUND=$(echo "$CONFIG" | cut -d'|' -f2)

    # Combine session name with notification type
    TITLE="$FORMATTED_SESSION - $TYPE_TITLE"

    # Check environment settings
    ENABLE_NOTIFICATIONS="${CLAUDEX_NOTIFICATIONS_ENABLED:-true}"
    ENABLE_VOICE="${CLAUDEX_VOICE_ENABLED:-false}"

    # Send visual notification
    if [ "$ENABLE_NOTIFICATIONS" = "true" ]; then
        send_notification "$TITLE" "$MESSAGE" "$SOUND"
        log_message "Notification sent: $TITLE - $MESSAGE (sound: $SOUND)"
    else
        log_message "Notifications disabled (CLAUDEX_NOTIFICATIONS_ENABLED=false)"
    fi

    # Send voice notification (if enabled)
    if [ "$ENABLE_VOICE" = "true" ]; then
        speak_message "$MESSAGE"
        log_message "Voice message sent: $MESSAGE"
    fi
else
    log_message "ERROR: Notification library not found: $NOTIFICATION_LIB"
    exit 1
fi

log_message "Hook completed successfully"
exit 0
