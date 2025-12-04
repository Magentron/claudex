// Package setup provides initialization and setup functionality for Claudex.
// It handles .claude directory setup, environment configuration, and
// file deployment for hooks, agents, and stacks.
package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"claudex/internal/fsutil"

	"github.com/spf13/afero"
)

// Environment abstracts environment variable access
type Environment interface {
	Get(key string) string
	Set(key, value string)
}

// HandleExistingClaudeDirectory checks if .claude exists and handles user choice
func HandleExistingClaudeDirectory(projectDir, claudeDir string) (proceed bool, err error) {
	// Silent merge: always proceed with setup
	return true, nil
}

// EnsureClaudeDirectoryWithDeps sets up the .claude directory in the project with injected dependencies
func EnsureClaudeDirectoryWithDeps(fs afero.Fs, env Environment, projectDir string, noOverwrite bool) error {
	claudeDir := filepath.Join(projectDir, ".claude")

	// Handle existing .claude directory with user choice
	proceed, err := HandleExistingClaudeDirectory(projectDir, claudeDir)
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("installation cancelled by user")
	}

	// Get config dir (~/.config/claudex)
	configDir := env.Get("XDG_CONFIG_HOME")
	if configDir == "" {
		home := env.Get("HOME")
		if home == "" {
			return fmt.Errorf("HOME environment variable not set")
		}
		configDir = filepath.Join(home, ".config")
	}
	claudexConfigDir := filepath.Join(configDir, "claudex")

	// Check if claudex config exists
	if _, err := fs.Stat(claudexConfigDir); err != nil {
		return fmt.Errorf("claudex config directory not found at %s - please run 'make install' first", claudexConfigDir)
	}

	// Create .claude directory structure
	hooksDir := filepath.Join(claudeDir, "hooks")
	agentsDir := filepath.Join(claudeDir, "agents")
	commandsAgentsDir := filepath.Join(claudeDir, "commands", "agents")

	if err := fs.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}
	if err := fs.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}
	if err := fs.MkdirAll(commandsAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands/agents directory: %w", err)
	}

	// Copy hooks from ~/.config/claudex/hooks/
	sourceHooksDir := filepath.Join(claudexConfigDir, "hooks")
	if _, err := fs.Stat(sourceHooksDir); err == nil {
		if err := fsutil.CopyDir(fs, sourceHooksDir, hooksDir, noOverwrite); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to copy hooks: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Hooks directory not found at %s\n", sourceHooksDir)
	}

	// Copy agent profiles to both agents/ and commands/agents/
	sourceAgentsDir := filepath.Join(claudexConfigDir, "profiles", "agents")
	if _, err := fs.Stat(sourceAgentsDir); err == nil {
		entries, err := afero.ReadDir(fs, sourceAgentsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not read agents directory: %v\n", err)
		} else {
			for _, entry := range entries {
				if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					sourcePath := filepath.Join(sourceAgentsDir, entry.Name())
					content, err := afero.ReadFile(fs, sourcePath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to read %s: %v\n", entry.Name(), err)
						continue
					}

					// Copy to agents/
					agentTarget := filepath.Join(agentsDir, entry.Name()+".md")
					if noOverwrite {
						if _, err := fs.Stat(agentTarget); err != nil {
							// File doesn't exist, write it
							if err := afero.WriteFile(fs, agentTarget, content, 0644); err != nil {
								fmt.Fprintf(os.Stderr, "Warning: Failed to copy to agents/%s: %v\n", entry.Name(), err)
							}
						}
					} else {
						if err := afero.WriteFile(fs, agentTarget, content, 0644); err != nil {
							fmt.Fprintf(os.Stderr, "Warning: Failed to copy to agents/%s: %v\n", entry.Name(), err)
						}
					}

					// Copy to commands/agents/
					commandTarget := filepath.Join(commandsAgentsDir, entry.Name()+".md")
					if noOverwrite {
						if _, err := fs.Stat(commandTarget); err != nil {
							// File doesn't exist, write it
							if err := afero.WriteFile(fs, commandTarget, content, 0644); err != nil {
								fmt.Fprintf(os.Stderr, "Warning: Failed to copy to commands/agents/%s: %v\n", entry.Name(), err)
							}
						}
					} else {
						if err := afero.WriteFile(fs, commandTarget, content, 0644); err != nil {
							fmt.Fprintf(os.Stderr, "Warning: Failed to copy to commands/agents/%s: %v\n", entry.Name(), err)
						}
					}
				}
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Profiles directory not found at %s\n", sourceAgentsDir)
	}

	// Generate settings.local.json
	settingsPath := filepath.Join(claudeDir, "settings.local.json")
	settingsContent := `{
  "permissions": {
    "allow": [],
    "deny": [],
    "ask": []
  },
  "hooks": {
    "Notification": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/notification-hook.sh"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/session-end.sh"
          }
        ]
      }
    ],
    "SubagentStop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/subagent-stop.sh"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "^(?!AskUserQuestion$).*",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/pre-tool-use.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-tool-use.sh"
          },
          {
            "type": "command",
            "command": ".claude/hooks/auto-doc-updater.sh"
          }
        ]
      }
    ]
  }
}
`
	// Check if noOverwrite and file exists
	if noOverwrite {
		if _, err := fs.Stat(settingsPath); err == nil {
			// File exists, skip writing
			goto skipSettings
		}
	}

	if err := afero.WriteFile(fs, settingsPath, []byte(settingsContent), 0644); err != nil {
		return fmt.Errorf("failed to write settings.local.json: %w", err)
	}

skipSettings:

	// Detect project stack and generate principal-engineer agents
	stacks := DetectProjectStacksWithFs(fs, projectDir)
	if len(stacks) == 0 {
		// Default to all stacks if none detected
		stacks = []string{"typescript", "python", "go"}
	}

	// Generate principal-engineer-{stack} agents
	rolesDir := filepath.Join(claudexConfigDir, "profiles", "roles")
	skillsDir := filepath.Join(claudexConfigDir, "profiles", "skills")

	for _, stack := range stacks {
		if err := AssembleEngineerAgentWithFs(fs, stack, agentsDir, commandsAgentsDir, rolesDir, skillsDir, noOverwrite); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to assemble principal-engineer-%s: %v\n", stack, err)
		}
	}

	// Create principal-engineer alias by copying the primary stack's agent
	if len(stacks) > 0 {
		primaryStack := stacks[0]
		aliasSource := filepath.Join(agentsDir, fmt.Sprintf("principal-engineer-%s.md", primaryStack))

		// Read the primary engineer content
		if aliasContent, err := afero.ReadFile(fs, aliasSource); err == nil {
			// Copy to agents/principal-engineer.md
			aliasAgentTarget := filepath.Join(agentsDir, "principal-engineer.md")
			if noOverwrite {
				if _, err := fs.Stat(aliasAgentTarget); err != nil {
					// File doesn't exist, write it
					if err := afero.WriteFile(fs, aliasAgentTarget, aliasContent, 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to create principal-engineer alias: %v\n", err)
					}
				}
			} else {
				if err := afero.WriteFile(fs, aliasAgentTarget, aliasContent, 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to create principal-engineer alias: %v\n", err)
				}
			}

			// Copy to commands/agents/principal-engineer.md
			aliasCommandTarget := filepath.Join(commandsAgentsDir, "principal-engineer.md")
			if noOverwrite {
				if _, err := fs.Stat(aliasCommandTarget); err != nil {
					// File doesn't exist, write it
					if err := afero.WriteFile(fs, aliasCommandTarget, aliasContent, 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to create principal-engineer command alias: %v\n", err)
					}
				}
			} else {
				if err := afero.WriteFile(fs, aliasCommandTarget, aliasContent, 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to create principal-engineer command alias: %v\n", err)
				}
			}
		}
	}

	fmt.Printf("✓ Created .claude directory with %d engineer profile(s)\n", len(stacks))
	return nil
}
