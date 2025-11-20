#!/usr/bin/env bash

set -e  # Exit on error
set -u  # Exit on undefined variable

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function for colored output
success() { echo -e "${GREEN}✓${NC} $1"; }
warning() { echo -e "${YELLOW}⚠${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
info() { echo -e "${BLUE}ℹ${NC} $1"; }

# Get the directory where this script is located (BMad installation directory)
BMAD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if target path is provided
if [ $# -eq 0 ]; then
    error "Error: Target project path is required"
    echo ""
    echo "Usage: $0 /path/to/your/project"
    echo ""
    echo "Example:"
    echo "  $0 /Users/username/my-project"
    echo ""
    exit 1
fi

TARGET_DIR="$1"

# Validate that target path exists
if [ ! -d "$TARGET_DIR" ]; then
    error "Error: Target directory does not exist: $TARGET_DIR"
    exit 1
fi

# Convert to absolute path
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"

info "BMad Uninstaller"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Target Dir: $TARGET_DIR"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

CLAUDE_PATH="$TARGET_DIR/.claude"
BACKUP_PATH="$TARGET_DIR/.claude.backup"

# Check if .claude exists
if [ ! -e "$CLAUDE_PATH" ]; then
    warning "No .claude directory found in target project"
    info "BMad may not be installed in this project"
    exit 0
fi

# Check if .claude is a symlink
if [ ! -L "$CLAUDE_PATH" ]; then
    error "Error: .claude exists but is not a symlink"
    warning "This may not be a BMad installation"
    echo ""
    echo "To protect your data, uninstall will not proceed."
    echo "If you want to remove it manually:"
    echo "  rm -rf $CLAUDE_PATH"
    echo ""
    exit 1
fi

# Get the symlink target
LINK_TARGET=$(readlink "$CLAUDE_PATH")

# Verify it points to a BMad installation (contains /.claude at the end)
if [[ ! "$LINK_TARGET" =~ /.claude$ ]]; then
    error "Error: .claude symlink does not point to a BMad installation"
    warning "Symlink target: $LINK_TARGET"
    echo ""
    echo "To protect your data, uninstall will not proceed."
    echo "If you want to remove it manually:"
    echo "  rm $CLAUDE_PATH"
    echo ""
    exit 1
fi

# Remove agent profile symbolic links first
info "Removing agent profile symbolic links..."

BMAD_CLAUDE=$(readlink "$CLAUDE_PATH")

# Define the agent profile links to remove (relative paths within BMad installation)
PROFILE_LINKS=(
    "agents/prompt-engineer.md"
    "commands/BMad/agents/prompt-engineer.md"
    "commands/BMad/agents/team-lead-new.md"
    "agents/researcher.md"
    "commands/BMad/agents/researcher.md"
    "agents/architect.md"
    "commands/BMad/agents/architect.md"
    "../.bmad-core/agents/architect.md"
    "../.bmad-infrastructure-devops/agents/infra-devops-platform.md"
    "commands/bmadInfraDevOps/agents/infra-devops-platform.md"
)

# Remove each symbolic link
for link_rel in "${PROFILE_LINKS[@]}"; do
    LINK_PATH="$BMAD_CLAUDE/$link_rel"
    
    # Check if it's a symbolic link and remove it
    if [ -L "$LINK_PATH" ]; then
        rm -f "$LINK_PATH"
    fi
done

success "Agent profile symbolic links removed"

# Remove claudex binary and profiles from PATH
info "Removing claudex binary from PATH..."

INSTALL_DIR="/usr/local/bin"
CLAUDEX_PATH="$INSTALL_DIR/claudex"
PROFILES_PATH="$INSTALL_DIR/.profiles"

if [ -L "$CLAUDEX_PATH" ] || [ -L "$PROFILES_PATH" ]; then
    # Check if we have write permission
    if [ -w "$INSTALL_DIR" ]; then
        [ -L "$CLAUDEX_PATH" ] && rm -f "$CLAUDEX_PATH"
        [ -L "$PROFILES_PATH" ] && rm -f "$PROFILES_PATH"
        success "Claudex binary and profiles removed"
    else
        # Need sudo for removal
        if sudo -v; then
            [ -L "$CLAUDEX_PATH" ] && sudo rm -f "$CLAUDEX_PATH"
            [ -L "$PROFILES_PATH" ] && sudo rm -f "$PROFILES_PATH"
            success "Claudex binary and profiles removed"
        else
            warning "Failed to remove claudex (requires sudo)"
            warning "You can manually remove them with:"
            warning "  sudo rm $CLAUDEX_PATH"
            warning "  sudo rm $PROFILES_PATH"
        fi
    fi
elif [ -e "$CLAUDEX_PATH" ]; then
    warning "Claudex exists but is not a symlink, skipping removal"
fi

# Remove the main .claude symlink
info "Removing BMad symlink..."
rm "$CLAUDE_PATH"
success "Symlink removed"
echo ""

# Check if backup exists and restore it
if [ -e "$BACKUP_PATH" ]; then
    info "Found backup: .claude.backup"
    info "Restoring original .claude directory..."
    mv "$BACKUP_PATH" "$CLAUDE_PATH"
    success "Original .claude directory restored"
    echo ""
    success "BMad uninstalled and original .claude restored!"
else
    success "BMad uninstalled successfully!"
    info "No backup found to restore"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Uninstallation Complete"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
