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
BMAD_CLAUDE="$BMAD_DIR/../.claude"

# Convert to absolute path and resolve symlinks
BMAD_CLAUDE="$(cd "$BMAD_CLAUDE" 2>/dev/null && pwd)"

if [ ! -d "$BMAD_CLAUDE" ]; then
    error "Error: BMad .claude directory not found at: $BMAD_DIR/../.claude"
    exit 1
fi

ln -s "$BMAD_CLAUDE" "$CLAUDE_PATH"

# Create symbolic links for agent profiles
info "Setting up agent profile symbolic links..."

PROFILES_DIR="$BMAD_DIR/../claudex-go/profiles"

# Ensure profiles directory exists
if [ ! -d "$PROFILES_DIR" ]; then
    error "Error: Profiles directory not found: $PROFILES_DIR"
    exit 1
fi

# Convert to absolute path
PROFILES_DIR="$(cd "$PROFILES_DIR" && pwd)"

# Define the symbolic links to create within .claude directory
# Format: "source_path:target_relative_to_bmad_claude"
PROFILE_LINKS_CLAUDE=(
    "$PROFILES_DIR/prompt-engineer:agents/prompt-engineer.md"
    "$PROFILES_DIR/prompt-engineer:commands/BMad/agents/prompt-engineer.md"
    "$PROFILES_DIR/team-lead-new:commands/BMad/agents/team-lead-new.md"
    "$PROFILES_DIR/researcher:agents/researcher.md"
    "$PROFILES_DIR/researcher:commands/BMad/agents/researcher.md"
    "$PROFILES_DIR/architect:agents/architect.md"
    "$PROFILES_DIR/architect:commands/BMad/agents/architect.md"
)

# Define symbolic links for other BMad directories (relative to BMAD_DIR)
PROFILE_LINKS_OTHER=(
    "$PROFILES_DIR/architect:../.bmad-core/agents/architect.md"
    "$PROFILES_DIR/infra-devops-platform:../.bmad-infrastructure-devops/agents/infra-devops-platform.md"
    "$PROFILES_DIR/infra-devops-platform:commands/bmadInfraDevOps/agents/infra-devops-platform.md"
)

# Create symbolic links within .claude directory
for link_spec in "${PROFILE_LINKS_CLAUDE[@]}"; do
    SOURCE="${link_spec%%:*}"
    TARGET_REL="${link_spec##*:}"
    TARGET="$BMAD_CLAUDE/$TARGET_REL"
    
    # Check if source exists
    if [ ! -f "$SOURCE" ]; then
        warning "Source profile not found: $SOURCE (skipping)"
        continue
    fi
    
    # Remove existing file/link if it exists
    if [ -e "$TARGET" ] || [ -L "$TARGET" ]; then
        rm -f "$TARGET"
    fi
    
    # Create parent directory if needed
    mkdir -p "$(dirname "$TARGET")"
    
    # Create symbolic link
    ln -s "$SOURCE" "$TARGET"
done

# Create symbolic links in other BMad directories
for link_spec in "${PROFILE_LINKS_OTHER[@]}"; do
    SOURCE="${link_spec%%:*}"
    TARGET_REL="${link_spec##*:}"
    TARGET="$BMAD_CLAUDE/$TARGET_REL"
    
    # Check if source exists
    if [ ! -f "$SOURCE" ]; then
        warning "Source profile not found: $SOURCE (skipping)"
        continue
    fi
    
    # Remove existing file/link if it exists
    if [ -e "$TARGET" ] || [ -L "$TARGET" ]; then
        rm -f "$TARGET"
    fi
    
    # Create parent directory if needed
    mkdir -p "$(dirname "$TARGET")"
    
    # Create symbolic link
    ln -s "$SOURCE" "$TARGET"
done

success "Agent profile symbolic links created"

# Install claudex binary to PATH
info "Installing claudex binary to PATH..."

CLAUDEX_BINARY="$BMAD_DIR/../claudex-go/claudex"
PROFILES_SOURCE="$BMAD_DIR/../claudex-go/profiles"
INSTALL_DIR="/usr/local/bin"

if [ ! -f "$CLAUDEX_BINARY" ]; then
    warning "Claudex binary not found: $CLAUDEX_BINARY (skipping)"
else
    # Check if we have write permission to /usr/local/bin
    if [ -w "$INSTALL_DIR" ]; then
        # Create symlink for claudex binary
        if [ -e "$INSTALL_DIR/claudex" ] || [ -L "$INSTALL_DIR/claudex" ]; then
            rm -f "$INSTALL_DIR/claudex"
        fi
        ln -s "$CLAUDEX_BINARY" "$INSTALL_DIR/claudex"
        
        # Create symlink for .profiles directory
        if [ -e "$INSTALL_DIR/.profiles" ] || [ -L "$INSTALL_DIR/.profiles" ]; then
            rm -f "$INSTALL_DIR/.profiles"
        fi
        ln -s "$PROFILES_SOURCE" "$INSTALL_DIR/.profiles"
        
        success "Claudex binary installed to $INSTALL_DIR/claudex"
        success "Profiles symlink created at $INSTALL_DIR/.profiles"
    else
        # Need sudo for installation
        echo ""
        warning "Installing claudex requires sudo privileges"
        if sudo -v; then
            if [ -e "$INSTALL_DIR/claudex" ] || [ -L "$INSTALL_DIR/claudex" ]; then
                sudo rm -f "$INSTALL_DIR/claudex"
            fi
            sudo ln -s "$CLAUDEX_BINARY" "$INSTALL_DIR/claudex"
            
            # Create symlink for .profiles directory
            if [ -e "$INSTALL_DIR/.profiles" ] || [ -L "$INSTALL_DIR/.profiles" ]; then
                sudo rm -f "$INSTALL_DIR/.profiles"
            fi
            sudo ln -s "$PROFILES_SOURCE" "$INSTALL_DIR/.profiles"
            
            success "Claudex binary installed to $INSTALL_DIR/claudex"
            success "Profiles symlink created at $INSTALL_DIR/.profiles"
        else
            error "Failed to get sudo privileges"
            warning "You can manually add claudex to your PATH by adding this to your shell config:"
            echo "    export PATH=\"\$PATH:$(dirname "$CLAUDEX_BINARY")\""
        fi
    fi
fi

echo ""
success "BMad installed successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  What Was Installed:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✓ BMad .claude directory symlinked to project"
echo "  ✓ Agent profile symlinks created (10 links)"
echo "  ✓ Claudex binary installed to /usr/local/bin"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Next Steps:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  1. Navigate to your project:"
echo "     cd $TARGET_DIR"
echo ""
echo "  2. Use claudex command (now available globally):"
echo "     claudex [command] [options]"
echo ""
echo "  3. Or use BMad agents with claude CLI:"
echo "     claude --system-prompt \"\$(cat .claude/commands/BMad/agents/team-lead.md)\" init"
echo ""
echo "  4. Or use the Makefile (if available in your project):"
echo "     make team-lead"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
