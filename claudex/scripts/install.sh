#!/usr/bin/env bash
set -e

# Claudex Installer
# Installs profiles and hooks to ~/.config/claudex
# Installs binary to ~/.local/bin

# Colors
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
success() { echo -e "${GREEN}✓${NC} $1"; }
warning() { echo -e "${YELLOW}⚠${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; exit 1; }

# Paths
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/claudex"
BIN_DIR="$HOME/.local/bin"
SCRIPT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Stack detection function (pure bash)
detect_stack() {
    local dir="${1:-.}"
    [[ -f "$dir/tsconfig.json" ]] && echo "typescript" && return 0
    [[ -f "$dir/go.mod" ]] && echo "go" && return 0
    [[ -f "$dir/pyproject.toml" || -f "$dir/requirements.txt" || -f "$dir/setup.py" || -f "$dir/Pipfile" ]] && echo "python" && return 0
    [[ -f "$dir/package.json" ]] && echo "javascript" && return 0
    echo "unknown" && return 1
}

# Agent assembly function
assemble_agent() {
    local stack="$1"
    local role="$SCRIPT_DIR/profiles/roles/engineer.md"
    local skill="$SCRIPT_DIR/profiles/skills/$stack.md"
    local out_dir="$CONFIG_DIR/generated"
    local out_file="$out_dir/principal-engineer-$stack"

    [[ -f "$role" ]] || error "Role template not found: $role"
    [[ -f "$skill" ]] || error "Skill file not found: $skill"

    mkdir -p "$out_dir"
    local stack_cap="$(tr '[:lower:]' '[:upper:]' <<< ${stack:0:1})${stack:1}"
    { sed "s/{Stack}/$stack_cap/g" "$role"; echo ""; cat "$skill"; } > "$out_file"
    success "Assembled principal-engineer-$stack"
}

# Main installation
main() {
    echo "Claudex Installer"
    echo "================="
    echo ""

    # Detect stack
    local stack=$(detect_stack ".")
    if [[ "$stack" == "unknown" ]]; then
        warning "No stack detected in current directory"
        echo "Supported: TypeScript, Go, Python, JavaScript"
    else
        success "Detected stack: $stack"
    fi

    # Create config directory
    echo ""
    echo "Installing to: $CONFIG_DIR"
    mkdir -p "$CONFIG_DIR"

    # Copy profiles
    [[ -d "$SCRIPT_DIR/profiles" ]] || error "Profiles directory not found"
    cp -r "$SCRIPT_DIR/profiles" "$CONFIG_DIR/"
    success "Installed profiles"

    # Copy hooks if they exist
    if [[ -d "$SCRIPT_DIR/.claude/hooks" ]]; then
        mkdir -p "$CONFIG_DIR/hooks"
        cp -r "$SCRIPT_DIR/.claude/hooks/"* "$CONFIG_DIR/hooks/"
        success "Installed hooks"
    fi

    # Assemble agents
    echo ""
    echo "Assembling dynamic agents..."
    mkdir -p "$CONFIG_DIR/generated"
    for skill in "$SCRIPT_DIR/profiles/skills"/*.md; do
        [[ -f "$skill" ]] || continue
        local name=$(basename "$skill" .md)
        [[ "$name" == "prompt-engineering" ]] && continue
        assemble_agent "$name" || warning "Failed to assemble $name agent"
    done

    # Install binary
    echo ""
    if [[ -f "$SCRIPT_DIR/claudex" ]]; then
        mkdir -p "$BIN_DIR"
        cp "$SCRIPT_DIR/claudex" "$BIN_DIR/claudex"
        chmod +x "$BIN_DIR/claudex"
        success "Installed binary to $BIN_DIR/claudex"
    else
        warning "Binary not found (build with 'make build' first)"
    fi

    # Check PATH
    echo ""
    if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
        warning "$BIN_DIR is not in your PATH"
        echo ""
        echo "Add to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    else
        success "$BIN_DIR is in your PATH"
    fi

    # Complete
    echo ""
    echo "Installation complete!"
    echo ""
    echo "Configuration: $CONFIG_DIR"
    echo "Binary: $BIN_DIR/claudex"
    [[ "$stack" != "unknown" ]] && echo "" && echo "Available: /agents:principal-engineer-$stack"
}

main "$@"
