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

info "BMad Installer"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  BMad Dir:   $BMAD_DIR"
echo "  Target Dir: $TARGET_DIR"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if .claude already exists in target
CLAUDE_PATH="$TARGET_DIR/.claude"

if [ -e "$CLAUDE_PATH" ]; then
    warning "Found existing .claude in target directory"
    echo ""

    # Check if it's already a symlink to our BMad installation
    if [ -L "$CLAUDE_PATH" ]; then
        LINK_TARGET=$(readlink "$CLAUDE_PATH")
        EXPECTED_TARGET="$BMAD_DIR/.claude"

        if [ "$LINK_TARGET" = "$EXPECTED_TARGET" ]; then
            success "BMad is already installed in this project!"
            exit 0
        else
            warning "Existing symlink points to: $LINK_TARGET"
        fi
    fi

    echo "How would you like to proceed?"
    echo "  [y] Backup existing .claude to .claude.backup and install BMad"
    echo "  [n] Overwrite existing .claude (no backup) and install BMad"
    echo "  [c] Cancel installation"
    echo ""

    while true; do
        read -p "Your choice (y/n/c): " choice
        case $choice in
            [Yy]* )
                BACKUP_PATH="$TARGET_DIR/.claude.backup"

                # Remove old backup if it exists
                if [ -e "$BACKUP_PATH" ]; then
                    warning "Removing old backup: $BACKUP_PATH"
                    rm -rf "$BACKUP_PATH"
                fi

                info "Creating backup: .claude.backup"
                mv "$CLAUDE_PATH" "$BACKUP_PATH"
                success "Backup created"
                break
                ;;
            [Nn]* )
                warning "Removing existing .claude without backup"
                rm -rf "$CLAUDE_PATH"
                success "Removed existing .claude"
                break
                ;;
            [Cc]* )
                info "Installation cancelled by user"
                exit 0
                ;;
            * )
                error "Invalid choice. Please enter y, n, or c."
                ;;
        esac
    done
    echo ""
fi

# Create the symlink
info "Creating symlink..."
BMAD_CLAUDE="$BMAD_DIR/.claude"

if [ ! -d "$BMAD_CLAUDE" ]; then
    error "Error: BMad .claude directory not found: $BMAD_CLAUDE"
    exit 1
fi

ln -s "$BMAD_CLAUDE" "$CLAUDE_PATH"

echo ""
success "BMad installed successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Next Steps:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  1. Navigate to your project:"
echo "     cd $TARGET_DIR"
echo ""
echo "  2. Use BMad agents:"
echo "     claude --system-prompt \"\$(cat .claude/commands/BMad/agents/team-lead.md)\" init"
echo ""
echo "  3. Or use the Makefile (if available in your project):"
echo "     make team-lead"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
