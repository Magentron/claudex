# BMad - Business Mad Science™

BMad is a comprehensive AI-powered development toolkit featuring intelligent agents, session management, and workflow automation for strategic planning, architecture design, and engineering execution.

## 🚀 Quick Start

### Prerequisites

- [Claude CLI](https://docs.anthropic.com/claude/docs/cli) installed and authenticated
- macOS (currently supported)
- Go 1.21+ (for building from source)

### Installation

**1. Clone the repository:**

```bash
git clone https://github.com/your-org/bmad.git
cd bmad
```

**2. Install BMad into your project:**

```bash
cd claudex-install
./install.sh /path/to/your/project
```

This will:
- ✅ Symlink BMad's `.claude` directory to your project
- ✅ Install `claudex` binary to `/usr/local/bin` for global access
- ✅ Set up agent profile symbolic links for single-source management

**Note:** Installation may require sudo privileges for system-wide binary installation.

### Using Claudex

After installation, you can launch `claudex` from anywhere:

```bash
# Navigate to any project directory
cd /path/to/your/project

# Launch claudex session manager
claudex
```

Claudex provides:
- 📋 **Session Management** - Create, resume, and fork working sessions
- 🎯 **Profile Selection** - Choose from multiple AI agent profiles
- 💾 **Session Persistence** - Save and resume your work across sessions
- 🔄 **Session Forking** - Branch off from existing sessions

## 📦 What's Included

### Claudex - Session Manager

A powerful terminal UI for managing Claude AI sessions with persistent workspace management.

**Features:**
- Create new sessions with AI-generated names
- Resume existing sessions with full context
- Fork sessions to explore alternative approaches
- Ephemeral mode for temporary work
- Profile-based agent activation
- Automatic session tracking and organization

**Usage:**
```bash
claudex                    # Launch interactive session manager
```

**Session Types:**
- **New Session**: Fresh start with selected profile
- **Resume Session**: Continue where you left off
- **Fork Session**: Branch from existing session with copied files
- **Ephemeral**: Work without saving session data

### BMad Agents

Intelligent AI agents with specialized roles and capabilities:

#### 🎯 **Team Lead Agent**
Strategic planning and team orchestration
- Create execution plans
- Delegate to engineering teams
- Coordinate infrastructure design
- Manage project workflows

#### 🏗️ **Principal Architect Agent**
System design and architecture decisions
- Gather requirements interactively
- Query up-to-date documentation (Context7)
- Design solution architectures
- Create implementation plans
- Optimize for parallel execution

#### 🔬 **Researcher Agent**
Deep analysis and investigation
- Codebase investigation
- Technology research with current docs
- Sequential thinking for complex analysis
- Evidence-based recommendations

#### 💻 **Principal TypeScript Engineer**
Implementation and code execution
- Execute plans from architect
- Build features with best practices
- Write comprehensive tests
- Handle implementation details

#### 🏗️ **Infrastructure/DevOps Agent**
Platform and deployment design
- Cloud-native architectures
- CI/CD pipeline design
- Infrastructure as Code
- Security and compliance

#### ✍️ **Prompt Engineer Agent**
Prompt design and optimization
- Craft effective prompts
- Optimize agent behaviors
- Test and iterate on prompt designs

## 🛠️ Building from Source

### Build Claudex Binary

```bash
cd claudex-go
make build
```

This creates the `claudex` binary in the `claudex-go` directory.

### Install Locally

```bash
cd claudex-go
make install
```

This copies the binary to `/usr/local/bin`.

## 📖 Usage Guide

### Method 1: Using Claudex (Recommended)

```bash
# From any directory
claudex

# Select session type:
# - Create New Session
# - Resume existing session
# - Fork session
# - Ephemeral mode

# Choose your agent profile
# Start working!
```

### Method 2: Direct Agent Activation

```bash
# Using Make (from BMad directory)
cd bmad
make team-lead

# Or using claude CLI directly
cd /path/to/your/project
claude --system-prompt "$(cat .claude/commands/BMad/agents/team-lead.md)" init
```

### Method 3: Profile Selector Script

```bash
# From BMad directory
./bmad.sh

# Interactive menu shows all available profiles
# Select the number of the agent you want
```

## 🗂️ Project Structure

```
bmad/
├── claudex-go/                 # Claudex session manager (Go)
│   ├── claudex                 # Compiled binary
│   ├── main.go                 # Main application
│   ├── ui.go                   # Terminal UI components
│   └── profiles/               # Agent profiles (master copies)
│       ├── architect
│       ├── researcher
│       ├── prompt-engineer
│       ├── team-lead-new
│       └── infra-devops-platform
│
├── sessions/                   # Your project sessions (created per-project)
│   └── [created in $(pwd)/sessions/ when you run claudex]
│
├── claudex-go-proxy/           # Development proxy (Go)
│   └── main.go
│
├── claudex-install/            # Installation scripts
│   ├── install.sh              # Install BMad to project
│   ├── uninstall.sh            # Remove BMad from project
│   └── README.md               # Installation documentation
│
├── .claude/                    # Claude AI configuration
│   ├── agents/                 # Agent definitions (symlinked)
│   └── commands/               # Agent commands (symlinked)
│
├── .bmad-core/                 # Core tasks and templates
│   └── agents/
│
├── .bmad-infrastructure-devops/  # Infrastructure agents
│   └── agents/
│
├── docs/                       # Documentation
│   ├── architecture/
│   └── product-definition.md
│
└── README.md                   # This file
```

## 🔧 Configuration

### Agent Profiles

Agent profiles are stored in `claudex-go/profiles/` as the single source of truth. The installation script creates symbolic links to these profiles in multiple locations for compatibility.

**To edit an agent:**
```bash
# Edit the master profile
vim claudex-go/profiles/architect

# Changes are automatically reflected everywhere via symlinks
```

### Session Storage

**Important:** Sessions are created in a `sessions/` directory within your **current working directory** when you launch `claudex`.

For example:
```bash
cd /path/to/my-project
claudex
# Creates: /path/to/my-project/sessions/
```

Each session directory contains:
- `.description` - Session description
- `.created` - Creation timestamp
- `.last_used` - Last access timestamp
- Session-specific files and data

**Best Practice:** Run `claudex` from your project root directory so sessions are organized per-project.

## 🚮 Uninstallation

To remove BMad from a project:

```bash
cd bmad/claudex-install
./uninstall.sh /path/to/your/project
```

This will:
1. Remove all agent profile symbolic links
2. Remove the `claudex` binary from `/usr/local/bin`
3. Remove the `.profiles` symlink from `/usr/local/bin`
4. Remove the BMad `.claude` symlink from your project
5. Restore your original `.claude` directory if a backup exists

**Note:** Uninstallation may require sudo privileges.

## 🐛 Troubleshooting

### Claudex Not Found

If `claudex` command is not found after installation:

```bash
# Check if binary exists
ls -la /usr/local/bin/claudex

# Check if /usr/local/bin is in PATH
echo $PATH | grep /usr/local/bin

# Add to PATH if needed (add to ~/.zshrc or ~/.bashrc)
export PATH="/usr/local/bin:$PATH"
```

### Permission Errors

Installation requires sudo for system-wide installation:

```bash
# The installer will prompt for sudo when needed
# Or manually install:
cd claudex-go
sudo ln -s "$(pwd)/claudex" /usr/local/bin/claudex
sudo ln -s "$(pwd)/profiles" /usr/local/bin/.profiles
```

### Profiles Not Found

If claudex can't find profiles:

```bash
# Check if .profiles symlink exists
ls -la /usr/local/bin/.profiles

# Should point to: /path/to/bmad/claudex-go/profiles

# Recreate if missing:
sudo ln -s /path/to/bmad/claudex-go/profiles /usr/local/bin/.profiles
```

### Sessions Not Saving

Sessions are saved in `./sessions/` relative to your current working directory:

```bash
# If you run claudex from here:
cd /path/to/your/project
claudex

# Sessions will be created in:
# /path/to/your/project/sessions/

# Each project can have its own sessions folder
cd ~/project-a && claudex  # Creates ~/project-a/sessions/
cd ~/project-b && claudex  # Creates ~/project-b/sessions/
```

**Tip:** Always run `claudex` from your project root to keep sessions organized.

## 🤝 Contributing

BMad is designed to be extended and customized:

1. **Add Custom Agents**: Create new profiles in `claudex-go/profiles/`
2. **Extend Workflows**: Add tasks and templates to `.bmad-custom/`
3. **Improve Core**: Submit PRs for core functionality improvements

## 📝 Development

### Prerequisites for Development

- Go 1.21+
- Make
- Claude CLI

### Development Workflow

```bash
# Build claudex
cd claudex-go
make build

# Run locally (without installing)
./claudex

# Install for testing
make install

# Clean build artifacts
make clean
```

### Testing Changes

```bash
# Test installation
cd claudex-install
./install.sh /path/to/test/project

# Test uninstallation
./uninstall.sh /path/to/test/project
```

## 📚 Additional Resources

- **Installation Guide**: See `claudex-install/README.md`
- **Claudex Documentation**: See `claudex-go/README.md`
- **Architecture Docs**: See `docs/architecture/`
- **Product Definition**: See `docs/product-definition.md`

## 📄 License

[Add your license information here]

## 🙏 Acknowledgments

Built with:
- [Claude AI](https://anthropic.com/claude) by Anthropic
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for terminal UI
- [Charm](https://charm.sh/) terminal libraries

---

**BMad** - Because coding should be as mad as your ideas! 🧪✨

