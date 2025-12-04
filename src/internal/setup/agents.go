package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// AssembleEngineerAgentWithFs creates a principal-engineer-{stack} agent from role + skill
func AssembleEngineerAgentWithFs(fs afero.Fs, stack, agentsDir, commandsAgentsDir, rolesDir, skillsDir string, noOverwrite bool) error {
	roleFile := filepath.Join(rolesDir, "engineer.md")
	skillFile := filepath.Join(skillsDir, stack+".md")

	// Read role template
	roleContent, err := afero.ReadFile(fs, roleFile)
	if err != nil {
		return fmt.Errorf("failed to read role file: %w", err)
	}

	// Capitalize stack name for display
	stackDisplay := strings.Title(stack)
	if stack == "typescript" {
		stackDisplay = "TypeScript"
	} else if stack == "go" {
		stackDisplay = "Go"
	}

	// Generate frontmatter
	frontmatter := fmt.Sprintf(`---
name: principal-engineer-%s
Description: Use this agent when you need a Principal %s Engineer for code implementation, debugging, refactoring, and development best practices. This agent executes stories by reading execution plans and implementing tasks sequentially with comprehensive testing and documentation lookup.
model: sonnet
color: blue
---

`, stack, stackDisplay)

	// Replace {Stack} placeholder in role content
	roleStr := strings.ReplaceAll(string(roleContent), "{Stack}", stackDisplay)

	// Read skill content if it exists
	var skillStr string
	if skillContent, err := afero.ReadFile(fs, skillFile); err == nil {
		skillStr = "\n" + string(skillContent)
	}

	// Combine all parts
	agentContent := frontmatter + roleStr + skillStr

	// Write to agents/ directory
	agentPath := filepath.Join(agentsDir, fmt.Sprintf("principal-engineer-%s.md", stack))
	if noOverwrite {
		if _, err := fs.Stat(agentPath); err != nil {
			// File doesn't exist, write it
			if err := afero.WriteFile(fs, agentPath, []byte(agentContent), 0644); err != nil {
				return fmt.Errorf("failed to write agent file: %w", err)
			}
		}
	} else {
		if err := afero.WriteFile(fs, agentPath, []byte(agentContent), 0644); err != nil {
			return fmt.Errorf("failed to write agent file: %w", err)
		}
	}

	// Copy to commands/agents/
	commandPath := filepath.Join(commandsAgentsDir, fmt.Sprintf("principal-engineer-%s.md", stack))
	if noOverwrite {
		if _, err := fs.Stat(commandPath); err != nil {
			// File doesn't exist, write it
			if err := afero.WriteFile(fs, commandPath, []byte(agentContent), 0644); err != nil {
				return fmt.Errorf("failed to write command file: %w", err)
			}
		}
	} else {
		if err := afero.WriteFile(fs, commandPath, []byte(agentContent), 0644); err != nil {
			return fmt.Errorf("failed to write command file: %w", err)
		}
	}

	return nil
}
