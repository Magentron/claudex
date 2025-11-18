# BMad - Business Mad Science™

BMad is a collection of AI agents and workflows for strategic planning, architecture design, and engineering execution.

## Installation

BMad can be installed into your projects to give you access to powerful AI agents like the Principal Team Lead, Principal Architect, and Principal TypeScript Engineer.

### Quick Install (Using Make)

The fastest way to install BMad into your project:

```bash
make install PROJECT_PATH=/path/to/your/project
```

### Manual Install (Using Script)

Alternatively, you can use the standalone installation script:

```bash
./install.sh /path/to/your/project
```

### What Gets Installed

The installer creates a symbolic link from BMad's `.claude` directory to your project. This gives you access to:

- **Team Lead Agent** - Strategic planning and team orchestration
- **Principal Architect** - System design and architecture decisions
- **Principal TypeScript Engineer** - Implementation and code execution
- **Infrastructure/DevOps Agent** - Platform and deployment design

### Handling Existing `.claude` Directories

If your project already has a `.claude` directory, the installer will prompt you with options:

- **[y] Backup** - Move existing `.claude` to `.claude.backup` before installing
- **[n] Overwrite** - Replace existing `.claude` without backup
- **[c] Cancel** - Exit without making changes

## Usage

### Quick Start - Profile Selector

The easiest way to launch BMad agents is using the interactive profile selector:

```bash
# From the BMad directory
make bmad

# Or run the script directly
./bmad.sh
```

This will display an interactive menu of all available BMad profiles. Simply select the number of the agent you want to activate!

### Direct Agent Activation

You can also activate agents directly:

```bash
# Using Make (from BMad directory)
make team-lead

# Or using claude directly (from any project with BMad installed)
cd /path/to/your/project
claude --system-prompt "$(cat .claude/commands/BMad/agents/team-lead.md)" init
```

### Available Agent Commands

Once an agent is activated, you can use their commands:

**Team Lead Commands:**
- `*help` - Show available commands
- `*plan-execution` - Create a phased execution plan
- `*execute` - Delegate implementation to the engineering team
- `*infrastructure` - Delegate infrastructure design
- `*exit` - Deactivate the current agent

## Uninstallation

To remove BMad from a project:

```bash
make uninstall PROJECT_PATH=/path/to/your/project
```

Or using the script:

```bash
./uninstall.sh /path/to/your/project
```

The uninstaller will:
1. Remove the BMad symlink
2. Restore your original `.claude` directory if a backup exists

## Troubleshooting

### Permission Errors

If you encounter permission errors during installation:

```bash
# Make sure the scripts are executable
chmod +x install.sh uninstall.sh

# Ensure you have write permissions to the target directory
ls -la /path/to/your/project
```

### Symlink Already Exists

If BMad is already installed in your project, the installer will detect it and inform you:

```
✓ BMad is already installed in this project!
```

### Verifying Installation

Check that the symlink was created correctly:

```bash
ls -la /path/to/your/project/.claude
# Should show: .claude -> /path/to/bmad/.claude
```

## Project Structure

```
bmad/
├── .bmad-core/              # Core tasks, templates, and data
├── .bmad-custom/            # Custom extensions
├── .bmad-infrastructure-devops/  # Infrastructure agents
├── .claude/                 # Agent definitions and commands
├── install.sh               # Installation script
├── uninstall.sh            # Uninstallation script
├── Makefile                # Convenient Make targets
└── README.md               # This file
```

## Contributing

BMad is designed to be extended. Add your own agents and workflows to `.bmad-custom/` and share them with your team.

## License

[Add your license information here]
