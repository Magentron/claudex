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

# Define BMad source paths (absolute)
BMAD_CLAUDE_SRC="$BMAD_DIR/../.claude"
BMAD_CLAUDE_SRC="$(cd "$(dirname "$BMAD_CLAUDE_SRC")" && pwd)/$(basename "$BMAD_CLAUDE_SRC")"

BMAD_CURSOR_SRC="$BMAD_DIR/../.cursor"
BMAD_CURSOR_SRC="$(cd "$(dirname "$BMAD_CURSOR_SRC")" && pwd)/$(basename "$BMAD_CURSOR_SRC")"

# --- Check .claude in target ---
CLAUDE_PATH="$TARGET_DIR/.claude"
SKIP_CLAUDE_LINK=false

if [ -e "$CLAUDE_PATH" ]; then
    warning "Found existing .claude in target directory"
    echo ""

    # Check if it's already a symlink to our BMad installation
    if [ -L "$CLAUDE_PATH" ]; then
        LINK_TARGET=$(readlink "$CLAUDE_PATH")
        
        if [ "$LINK_TARGET" = "$BMAD_CLAUDE_SRC" ]; then
            success "BMad .claude is already correctly linked."
            SKIP_CLAUDE_LINK=true
        else
            warning "Existing symlink points to: $LINK_TARGET"
        fi
    fi

    if [ "$SKIP_CLAUDE_LINK" = false ]; then
        echo "How would you like to proceed with .claude?"
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
    fi
    echo ""
fi

# --- Check .cursor in target ---
CURSOR_PATH="$TARGET_DIR/.cursor"
SKIP_CURSOR_LINK=false

if [ -e "$CURSOR_PATH" ]; then
    warning "Found existing .cursor in target directory"
    echo ""

    # Check if it's already a symlink to our BMad installation
    if [ -L "$CURSOR_PATH" ]; then
        LINK_TARGET=$(readlink "$CURSOR_PATH")
        
        if [ "$LINK_TARGET" = "$BMAD_CURSOR_SRC" ]; then
            success "BMad .cursor is already correctly linked."
            SKIP_CURSOR_LINK=true
        else
            warning "Existing symlink points to: $LINK_TARGET"
        fi
    fi

    if [ "$SKIP_CURSOR_LINK" = false ]; then
        echo "How would you like to proceed with .cursor?"
        echo "  [y] Backup existing .cursor to .cursor.backup and install BMad"
        echo "  [n] Overwrite existing .cursor (no backup) and install BMad"
        echo "  [s] Skip .cursor installation"
        echo ""

        while true; do
            read -p "Your choice (y/n/s): " choice
            case $choice in
                [Yy]* )
                    BACKUP_PATH="$TARGET_DIR/.cursor.backup"
                    if [ -e "$BACKUP_PATH" ]; then
                         rm -rf "$BACKUP_PATH"
                    fi
                    info "Creating backup: .cursor.backup"
                    mv "$CURSOR_PATH" "$BACKUP_PATH"
                    success "Backup created"
                    break
                    ;;
                [Nn]* )
                    warning "Removing existing .cursor without backup"
                    rm -rf "$CURSOR_PATH"
                    break
                    ;;
                [Ss]* )
                    SKIP_CURSOR_LINK=true
                    info "Skipping .cursor installation"
                    break
                    ;;
                * )
                    error "Invalid choice."
                    ;;
            esac
        done
    fi
    echo ""
fi


# --- Set up BMad .claude directory ---
info "Setting up BMad .claude directory..."

# Create BMad .claude directory if it doesn't exist
if [ ! -d "$BMAD_CLAUDE_SRC" ]; then
    info "Creating BMad .claude directory: $BMAD_CLAUDE_SRC"
    mkdir -p "$BMAD_CLAUDE_SRC"
fi

# Source directory for claudex-go/.claude
CLAUDEX_CLAUDE_DIR="$BMAD_DIR/../claudex-go/.claude"

# Ensure claudex-go/.claude directory exists
if [ ! -d "$CLAUDEX_CLAUDE_DIR" ]; then
    error "Error: Claudex .claude directory not found: $CLAUDEX_CLAUDE_DIR"
    exit 1
fi

# Convert to absolute path
CLAUDEX_CLAUDE_DIR="$(cd "$CLAUDEX_CLAUDE_DIR" && pwd)"

# Create symbolic links for agent profiles
info "Setting up agent profile symbolic links..."

PROFILES_DIR="$BMAD_DIR/../claudex-go/profiles"
AGENTS_DIR="$PROFILES_DIR/agents"

# Ensure profiles directory exists
if [ ! -d "$PROFILES_DIR" ]; then
    error "Error: Profiles directory not found: $PROFILES_DIR"
    exit 1
fi

# Ensure agents directory exists
if [ ! -d "$AGENTS_DIR" ]; then
    error "Error: Agents directory not found: $AGENTS_DIR"
    exit 1
fi

# Convert to absolute paths
PROFILES_DIR="$(cd "$PROFILES_DIR" && pwd)"
AGENTS_DIR="$(cd "$AGENTS_DIR" && pwd)"

# Create symbolic link to the profiles directory itself in .claude
info "Creating profiles directory symlink in .claude..."
PROFILES_LINK_TARGET="$BMAD_CLAUDE_SRC/profiles"
if [ -e "$PROFILES_LINK_TARGET" ] || [ -L "$PROFILES_LINK_TARGET" ]; then
    rm -f "$PROFILES_LINK_TARGET"
fi
ln -s "$PROFILES_DIR" "$PROFILES_LINK_TARGET"

# Populate BMad .claude with symlinks from claudex-go/.claude
info "Populating BMad .claude with symlinks from claudex-go/.claude..."

while IFS= read -r -d '' file; do
    REL_PATH="${file#$CLAUDEX_CLAUDE_DIR/}"
    if [ "$file" = "$CLAUDEX_CLAUDE_DIR" ]; then continue; fi
    TARGET="$BMAD_CLAUDE_SRC/$REL_PATH"
    mkdir -p "$(dirname "$TARGET")"
    if [ -e "$TARGET" ] || [ -L "$TARGET" ]; then rm -f "$TARGET"; fi
    ln -s "$file" "$TARGET"
done < <(find "$CLAUDEX_CLAUDE_DIR" -type f -print0)

success "BMad .claude populated with symlinks"

# Dynamically create symbolic links for all agent profile files in .claude
info "Discovering and linking agent profiles in .claude..."

mkdir -p "$BMAD_CLAUDE_SRC/agents"
mkdir -p "$BMAD_CLAUDE_SRC/commands/agents"

while IFS= read -r -d '' profile_file; do
    PROFILE_NAME="$(basename "$profile_file")"
    if [ -d "$profile_file" ] || [[ "$PROFILE_NAME" == .* ]]; then continue; fi
    
    AGENT_TARGET="$BMAD_CLAUDE_SRC/agents/${PROFILE_NAME}.md"
    COMMAND_TARGET="$BMAD_CLAUDE_SRC/commands/agents/${PROFILE_NAME}.md"
    
    [ -e "$AGENT_TARGET" ] || [ -L "$AGENT_TARGET" ] && rm -f "$AGENT_TARGET"
    [ -e "$COMMAND_TARGET" ] || [ -L "$COMMAND_TARGET" ] && rm -f "$COMMAND_TARGET"
    
    ln -s "$profile_file" "$AGENT_TARGET"
    ln -s "$profile_file" "$COMMAND_TARGET"
    
    info "  Linked profile (claude): $PROFILE_NAME"
done < <(find "$AGENTS_DIR" -maxdepth 1 -type f -print0)

success "Agent profile symbolic links created in .claude"


# --- Set up BMad .cursor directory ---
info "Setting up BMad .cursor directory..."

if [ ! -d "$BMAD_CURSOR_SRC" ]; then
    mkdir -p "$BMAD_CURSOR_SRC"
fi

if [ ! -d "$BMAD_CURSOR_SRC/rules" ]; then
    mkdir -p "$BMAD_CURSOR_SRC/rules"
fi

info "Linking agent profiles to .cursor/rules..."
while IFS= read -r -d '' profile_file; do
    PROFILE_NAME="$(basename "$profile_file")"
    if [ -d "$profile_file" ] || [[ "$PROFILE_NAME" == .* ]]; then continue; fi
    
    # Link with .mdc extension
    RULE_TARGET="$BMAD_CURSOR_SRC/rules/${PROFILE_NAME}.mdc"
    
    [ -e "$RULE_TARGET" ] || [ -L "$RULE_TARGET" ] && rm -f "$RULE_TARGET"
    
    ln -s "$profile_file" "$RULE_TARGET"
    info "  Linked rule: ${PROFILE_NAME}.mdc"
done < <(find "$AGENTS_DIR" -maxdepth 1 -type f -print0)

success "Agent profiles linked to .cursor/rules"


# --- Link target .claude ---
if [ "$SKIP_CLAUDE_LINK" = false ]; then
    info "Creating symlink from target to BMad .claude..."
    ln -s "$BMAD_CLAUDE_SRC" "$CLAUDE_PATH"
    success "Target project .claude symlinked to BMad .claude"
else
    info "Skipping .claude link creation (already linked or skipped)"
fi

# --- Link target .cursor ---
if [ "$SKIP_CURSOR_LINK" = false ]; then
    info "Creating symlink from target to BMad .cursor..."
    ln -s "$BMAD_CURSOR_SRC" "$CURSOR_PATH"
    success "Target project .cursor symlinked to BMad .cursor"
else
    info "Skipping .cursor link creation (already linked or skipped)"
fi


# --- Install binary ---
info "Installing claudex binary to PATH..."

info "Building claudex binary..."
if ! (cd "$BMAD_DIR/../claudex-go" && make build >/dev/null 2>&1); then
    error "Failed to build claudex binary. Make sure Go is installed."
    warning "Skipping binary installation."
else
    success "Claudex binary built successfully"
fi

CLAUDEX_BINARY="$BMAD_DIR/../claudex-go/claudex"
INSTALL_DIR="/usr/local/bin"

if [ ! -f "$CLAUDEX_BINARY" ]; then
    warning "Claudex binary not found: $CLAUDEX_BINARY (skipping)"
else
    if [ -w "$INSTALL_DIR" ]; then
        if [ -e "$INSTALL_DIR/claudex" ] || [ -L "$INSTALL_DIR/claudex" ]; then
            rm -f "$INSTALL_DIR/claudex"
        fi
        ln -s "$CLAUDEX_BINARY" "$INSTALL_DIR/claudex"
        if [ -e "$INSTALL_DIR/.profiles" ] || [ -L "$INSTALL_DIR/.profiles" ]; then
            rm -f "$INSTALL_DIR/.profiles"
        fi
        success "Claudex binary installed to $INSTALL_DIR/claudex"
    else
        echo ""
        warning "Installing claudex requires sudo privileges"
        if sudo -v; then
            if [ -e "$INSTALL_DIR/claudex" ] || [ -L "$INSTALL_DIR/claudex" ]; then
                sudo rm -f "$INSTALL_DIR/claudex"
            fi
            sudo ln -s "$CLAUDEX_BINARY" "$INSTALL_DIR/claudex"
            if [ -e "$INSTALL_DIR/.profiles" ] || [ -L "$INSTALL_DIR/.profiles" ]; then
                sudo rm -f "$INSTALL_DIR/.profiles"
            fi
            success "Claudex binary installed to $INSTALL_DIR/claudex"
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
echo "  ✓ BMad .claude directory created/updated"
echo "  ✓ BMad .claude populated with claudex-go/.claude symlinks"
echo "  ✓ Agent profile symlinks created in BMad .claude"
echo "  ✓ Agent profile symlinks created in BMad .cursor/rules"
echo "  ✓ Profiles directory symlinked in BMad .claude"
if [ "$SKIP_CLAUDE_LINK" = false ]; then
    echo "  ✓ Target project .claude symlinked to BMad .claude"
fi
if [ "$SKIP_CURSOR_LINK" = false ]; then
    echo "  ✓ Target project .cursor symlinked to BMad .cursor"
fi
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
echo "     claude --system-prompt \"\$(cat .claude/commands/agents/team-lead-new.md)\" init"
echo ""
echo "  4. Or use Cursor with the installed rules."
echo ""
