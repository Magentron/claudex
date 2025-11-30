package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

//go:embed profiles
var profilesFS embed.FS

// ensureClaudeDirectory sets up the .claude directory in the project (always overwrites)
func ensureClaudeDirectory(projectDir string) error {
	claudeDir := filepath.Join(projectDir, ".claude")

	// Remove existing .claude if it exists (file or directory) - always overwrite
	if _, err := os.Stat(claudeDir); err == nil {
		if err := os.RemoveAll(claudeDir); err != nil {
			return fmt.Errorf("failed to remove existing .claude: %w", err)
		}
	}

	// Get config dir (~/.config/claudex)
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return fmt.Errorf("HOME environment variable not set")
		}
		configDir = filepath.Join(home, ".config")
	}
	claudexConfigDir := filepath.Join(configDir, "claudex")

	// Check if claudex config exists
	if _, err := os.Stat(claudexConfigDir); os.IsNotExist(err) {
		return fmt.Errorf("claudex config directory not found at %s - please run 'make install' first", claudexConfigDir)
	}

	// Create .claude directory structure
	hooksDir := filepath.Join(claudeDir, "hooks")
	agentsDir := filepath.Join(claudeDir, "agents")
	commandsAgentsDir := filepath.Join(claudeDir, "commands", "agents")

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}
	if err := os.MkdirAll(commandsAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands/agents directory: %w", err)
	}

	// Copy hooks from ~/.config/claudex/hooks/
	sourceHooksDir := filepath.Join(claudexConfigDir, "hooks")
	if _, err := os.Stat(sourceHooksDir); err == nil {
		if err := copyDir(sourceHooksDir, hooksDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to copy hooks: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Hooks directory not found at %s\n", sourceHooksDir)
	}

	// Copy agent profiles to both agents/ and commands/agents/
	sourceAgentsDir := filepath.Join(claudexConfigDir, "profiles", "agents")
	if _, err := os.Stat(sourceAgentsDir); err == nil {
		entries, err := os.ReadDir(sourceAgentsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not read agents directory: %v\n", err)
		} else {
			for _, entry := range entries {
				if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					sourcePath := filepath.Join(sourceAgentsDir, entry.Name())
					content, err := os.ReadFile(sourcePath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to read %s: %v\n", entry.Name(), err)
						continue
					}

					// Copy to agents/
					agentTarget := filepath.Join(agentsDir, entry.Name()+".md")
					if err := os.WriteFile(agentTarget, content, 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to copy to agents/%s: %v\n", entry.Name(), err)
					}

					// Copy to commands/agents/
					commandTarget := filepath.Join(commandsAgentsDir, entry.Name()+".md")
					if err := os.WriteFile(commandTarget, content, 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to copy to commands/agents/%s: %v\n", entry.Name(), err)
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
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0644); err != nil {
		return fmt.Errorf("failed to write settings.local.json: %w", err)
	}

	// Detect project stack and generate principal-engineer agents
	stacks := detectProjectStacks(projectDir)
	if len(stacks) == 0 {
		// Default to all stacks if none detected
		stacks = []string{"typescript", "python", "go"}
	}

	// Generate principal-engineer-{stack} agents
	rolesDir := filepath.Join(claudexConfigDir, "profiles", "roles")
	skillsDir := filepath.Join(claudexConfigDir, "profiles", "skills")

	for _, stack := range stacks {
		if err := assembleEngineerAgent(stack, agentsDir, commandsAgentsDir, rolesDir, skillsDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to assemble principal-engineer-%s: %v\n", stack, err)
		}
	}

	// Create principal-engineer alias by copying the primary stack's agent
	if len(stacks) > 0 {
		primaryStack := stacks[0]
		aliasSource := filepath.Join(agentsDir, fmt.Sprintf("principal-engineer-%s.md", primaryStack))

		// Read the primary engineer content
		if aliasContent, err := os.ReadFile(aliasSource); err == nil {
			// Copy to agents/principal-engineer.md
			aliasAgentTarget := filepath.Join(agentsDir, "principal-engineer.md")
			if err := os.WriteFile(aliasAgentTarget, aliasContent, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to create principal-engineer alias: %v\n", err)
			}

			// Copy to commands/agents/principal-engineer.md
			aliasCommandTarget := filepath.Join(commandsAgentsDir, "principal-engineer.md")
			if err := os.WriteFile(aliasCommandTarget, aliasContent, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to create principal-engineer command alias: %v\n", err)
			}
		}
	}

	fmt.Printf("✓ Created .claude directory with %d engineer profile(s)\n", len(stacks))
	return nil
}

// detectProjectStacks detects technology stacks based on marker files (searches up to 3 levels deep)
func detectProjectStacks(projectDir string) []string {
	var stacks []string

	// TypeScript detection
	if findFile(projectDir, "tsconfig.json", 3) {
		stacks = append(stacks, "typescript")
	} else if findFile(projectDir, "package.json", 3) {
		stacks = append(stacks, "typescript")
	}

	// Go detection
	if findFile(projectDir, "go.mod", 3) {
		stacks = append(stacks, "go")
	}

	// Python detection
	if findFile(projectDir, "pyproject.toml", 3) ||
		findFile(projectDir, "requirements.txt", 3) ||
		findFile(projectDir, "setup.py", 3) ||
		findFile(projectDir, "Pipfile", 3) {
		stacks = append(stacks, "python")
	}

	return stacks
}

// findFile searches for a file in projectDir and subdirectories up to maxDepth
func findFile(dir string, filename string, maxDepth int) bool {
	if maxDepth < 0 {
		return false
	}

	// Check current directory
	if fileExists(filepath.Join(dir, filename)) {
		return true
	}

	// Search subdirectories
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			if findFile(filepath.Join(dir, entry.Name()), filename, maxDepth-1) {
				return true
			}
		}
	}

	return false
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// assembleEngineerAgent creates a principal-engineer-{stack} agent from role + skill
func assembleEngineerAgent(stack, agentsDir, commandsAgentsDir, rolesDir, skillsDir string) error {
	roleFile := filepath.Join(rolesDir, "engineer.md")
	skillFile := filepath.Join(skillsDir, stack+".md")

	// Read role template
	roleContent, err := os.ReadFile(roleFile)
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
description: Use this agent when you need a Principal %s Engineer for code implementation, debugging, refactoring, and development best practices. This agent executes stories by reading execution plans and implementing tasks sequentially with comprehensive testing and documentation lookup.
model: sonnet
color: blue
---

`, stack, stackDisplay)

	// Replace {Stack} placeholder in role content
	roleStr := strings.ReplaceAll(string(roleContent), "{Stack}", stackDisplay)

	// Read skill content if it exists
	var skillStr string
	if skillContent, err := os.ReadFile(skillFile); err == nil {
		skillStr = "\n" + string(skillContent)
	}

	// Combine all parts
	agentContent := frontmatter + roleStr + skillStr

	// Write to agents/ directory
	agentPath := filepath.Join(agentsDir, fmt.Sprintf("principal-engineer-%s.md", stack))
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	// Copy to commands/agents/
	commandPath := filepath.Join(commandsAgentsDir, fmt.Sprintf("principal-engineer-%s.md", stack))
	if err := os.WriteFile(commandPath, []byte(agentContent), 0644); err != nil {
		return fmt.Errorf("failed to write command file: %w", err)
	}

	return nil
}

// buildSystemPromptWithSessionContext injects session context into the system prompt
// to ensure all agents follow session folder documentation rules.
func buildSystemPromptWithSessionContext(profileContent []byte, sessionPath string) (string, error) {
	// Skip injection for ephemeral sessions (empty sessionPath)
	if sessionPath == "" {
		return string(profileContent), nil
	}

	// List files in session directory (excluding hidden files starting with '.')
	entries, err := os.ReadDir(sessionPath)
	if err != nil {
		return "", fmt.Errorf("failed to read session directory: %w", err)
	}

	// Build file listing
	var files []string
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files (starting with '.')
		if !strings.HasPrefix(name, ".") {
			files = append(files, name)
		}
	}

	var filesDisplay string
	if len(files) == 0 {
		filesDisplay = "- (No files yet - you'll be the first to create documentation!)"
	} else {
		// Format as bullet list
		for _, f := range files {
			filesDisplay += fmt.Sprintf("- %s\n", f)
		}
		filesDisplay = strings.TrimSuffix(filesDisplay, "\n")
	}

	// Build session context template (from pre-tool-use.sh hook)
	sessionContext := fmt.Sprintf(`## SESSION CONTEXT (CRITICAL)

You are working within an active Claudex session. ALL documentation, plans, and artifacts MUST be created in the session folder.

**Session Folder (Absolute Path)**: `+"`%s`"+`

### MANDATORY RULES for Documentation:
1. ✅ ALWAYS save documentation to the session folder above
2. ✅ Use absolute paths when creating files (Write/Edit tools)
3. ✅ Before exploring the codebase, check the session folder for existing context
4. ❌ NEVER save documentation to project root or arbitrary locations
5. ❌ NEVER use relative paths for documentation files

### Session Folder Contents:
%s

### Recommended File Names:
- Research documents: `+"`research-{topic}.md`"+`
- Execution plans: `+"`execution-plan-{feature}.md`"+`
- Analysis reports: `+"`analysis-{component}.md`"+`
- Technical specs: `+"`technical-spec-{feature}.md`"+`

---

`, sessionPath, filesDisplay)

	// Concatenate session context with profile content
	combinedPrompt := sessionContext + "\n" + string(profileContent)
	return combinedPrompt, nil
}

func main() {
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Ensure .claude directory is set up
	if err := ensureClaudeDirectory(projectDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up .claude directory: %v\n", err)
		os.Exit(1)
	}

	// Setup centralized logging
	logsDir := filepath.Join(projectDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create logs directory: %v\n", err)
	}

	// Create unique log file for this execution
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("claudex-%s.log", timestamp)
	logFilePath := filepath.Join(logsDir, logFileName)

	// Open log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not open log file: %v\n", err)
	} else {
		defer logFile.Close()
		// Configure Go logger with [claudex] prefix
		log.SetOutput(logFile)
		log.SetPrefix("[claudex] ")
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

		// Set environment variable for hooks
		os.Setenv("CLAUDEX_LOG_FILE", logFilePath)

		log.Printf("Claudex started (log file: %s)", logFileName)
	}

	sessionsDir := filepath.Join(projectDir, "sessions")

	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get sessions
	sessions, err := getSessions(sessionsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Build items
	items := []list.Item{
		sessionItem{title: "Create New Session", description: "Start a fresh working session", itemType: "new"},
		sessionItem{title: "Ephemeral", description: "Work without saving session data", itemType: "ephemeral"},
	}

	for _, s := range sessions {
		items = append(items, s)
	}

	// Create list
	delegate := itemDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = "Claudex Session Manager"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Additional keybindings
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("q"),
				key.WithHelp("q", "quit"),
			),
		}
	}

	m := model{
		list:        l,
		stage:       "session",
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := finalModel.(model)
	if fm.quitting {
		return
	}

	// Handle "Create New Session" - use team-lead profile directly
	var profileContent []byte
	if fm.choice == "new" {
		// Load team-lead profile directly (skip profile selection menu)
		var err error
		profileContent, err = loadComposedProfile("team-lead")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
			os.Exit(1)
		}

		// Create the session with team-lead profile
		sessionName, sessionPath, claudeSessionID, err := createNewSessionParallel(sessionsDir, profileContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fm.sessionName = sessionName
		fm.sessionPath = sessionPath
		fm.choice = claudeSessionID // Store session ID for later use
	}

	// Check if selected session has a Claude session ID (for resume/fork choice)
	var resumeOrForkChoice string
	var isFreshMemory bool // Track if "fresh memory" was chosen
	if fm.choice == "session" && hasClaudeSessionID(fm.sessionName) {
		// Show resume/fork menu
		resumeOrForkItems := []list.Item{
			sessionItem{title: "Resume Session", description: "Continue with existing context", itemType: "resume"},
			sessionItem{title: "Fork Session", description: "Start fresh with copied files", itemType: "fork"},
		}

		delegate := itemDelegate{}
		rfList := list.New(resumeOrForkItems, delegate, 0, 0)
		rfList.Title = fmt.Sprintf("Resume or Fork • Session: %s", fm.sessionName)
		rfList.Styles.Title = titleStyle
		rfList.SetShowStatusBar(false)
		rfList.SetFilteringEnabled(false)
		rfList.SetShowHelp(true)

		rfModel := model{
			list:        rfList,
			stage:       "resume_or_fork",
			sessionName: fm.sessionName,
			sessionPath: fm.sessionPath,
			projectDir:  projectDir,
			sessionsDir: sessionsDir,
		}

		rfProgram := tea.NewProgram(rfModel, tea.WithAltScreen())
		finalRfModel, err := rfProgram.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		rfm := finalRfModel.(model)
		if rfm.quitting {
			return
		}

		resumeOrForkChoice = rfm.choice

		// Add variable to track resume submenu choice
		var resumeSubmenuChoice string

		// If user chose "Resume Session", show submenu: Continue vs Fresh Memory
		if resumeOrForkChoice == "resume" {
			// Show resume submenu: Continue with context vs Fresh memory
			resumeSubmenuItems := []list.Item{
				sessionItem{title: "Continue with context", description: "Resume with full conversation history", itemType: "continue"},
				sessionItem{title: "Fresh memory", description: "Start fresh, keep files, delete original", itemType: "fresh"},
			}

			delegate := itemDelegate{}
			rsMenu := list.New(resumeSubmenuItems, delegate, 0, 0)
			rsMenu.Title = fmt.Sprintf("Resume Options • Session: %s", fm.sessionName)
			rsMenu.Styles.Title = titleStyle
			rsMenu.SetShowStatusBar(false)
			rsMenu.SetFilteringEnabled(false)
			rsMenu.SetShowHelp(true)

			rsModel := model{
				list:        rsMenu,
				stage:       "resume_submenu",
				sessionName: fm.sessionName,
				sessionPath: fm.sessionPath,
				projectDir:  projectDir,
				sessionsDir: sessionsDir,
			}

			rsProgram := tea.NewProgram(rsModel, tea.WithAltScreen())
			finalRsModel, err := rsProgram.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			rsm := finalRsModel.(model)
			if rsm.quitting {
				return
			}

			resumeSubmenuChoice = rsm.choice

			// Handle "Fresh Memory" choice
			if resumeSubmenuChoice == "fresh" {
				newSessionName, newSessionPath, newClaudeSessionID, err := freshMemorySession(sessionsDir, fm.sessionName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating fresh session: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("\n🔄 Fresh memory: %s → %s (original deleted)\n", fm.sessionName, newSessionName)
				fm.sessionName = newSessionName
				fm.sessionPath = newSessionPath
				fm.choice = newClaudeSessionID
				isFreshMemory = true        // Track that this is a fresh memory session
				resumeOrForkChoice = "fork" // Reuse fork launch path (--session-id)
			}
			// else: resumeSubmenuChoice == "continue" -> proceed with existing resume logic
		}

		// Handle fork choice (but not for fresh memory - already processed above)
		if resumeOrForkChoice == "fork" && !isFreshMemory {
			// Prompt for new description (similar to createNewSessionParallel)
			fmt.Print("\033[H\033[2J") // Clear screen
			fmt.Println()
			fmt.Println("\033[1;36m Fork Session \033[0m")
			fmt.Printf("  Original: %s\n", fm.sessionName)
			fmt.Println()

			fmt.Print("  Description for fork: ")
			reader := bufio.NewReader(os.Stdin)
			forkDescription, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading description: %v\n", err)
				os.Exit(1)
			}
			forkDescription = strings.TrimSpace(forkDescription)

			if forkDescription == "" {
				fmt.Fprintf(os.Stderr, "Error: description cannot be empty\n")
				os.Exit(1)
			}

			newSessionName, newSessionPath, newClaudeSessionID, err := forkSessionWithDescription(sessionsDir, fm.sessionName, forkDescription)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error forking session: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n✅ Forked session: %s → %s\n", fm.sessionName, newSessionName)
			fm.sessionName = newSessionName
			fm.sessionPath = newSessionPath
			fm.choice = newClaudeSessionID // Store the new session ID
		}
	}

	// Set environment
	os.Setenv("CLAUDEX_SESSION", fm.sessionName)
	os.Setenv("CLAUDEX_SESSION_PATH", fm.sessionPath)

	// Handle resume vs new/fork session
	var claudeSessionID string
	var isNewSessionAlreadyInitialized bool

	// Check if we just created a new session (session ID stored in fm.choice)
	if fm.choice != "new" && fm.choice != "session" && fm.choice != "ephemeral" && len(fm.choice) > 30 {
		// This is a Claude session ID from createNewSessionParallel
		claudeSessionID = fm.choice
		isNewSessionAlreadyInitialized = true
	}

	if isNewSessionAlreadyInitialized {
		// New session created, launch it with --session-id
		// Update last used timestamp
		if err := updateLastUsed(fm.sessionPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
		}

		// Give terminal a moment to settle
		time.Sleep(100 * time.Millisecond)

		// Clear screen and show launching message
		fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
		fmt.Print("\033[0m")              // Reset all attributes
		fmt.Printf("\n✅ Launching new Claude session\n")
		fmt.Printf("📦 Session: %s\n", fm.sessionName)
		fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		// Construct relative session path for activation command
		relativeSessionPath := filepath.Join("sessions", filepath.Base(fm.sessionPath))
		activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", relativeSessionPath)

		// Launch the Claude session with activation command
		launchCmd := exec.Command("claude", "--session-id", claudeSessionID, activationPrompt)
		launchCmd.Stdin = os.Stdin
		launchCmd.Stdout = os.Stdout
		launchCmd.Stderr = os.Stderr
		launchCmd.Env = os.Environ() // Ensure environment variables are inherited

		if err := launchCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
			os.Exit(1)
		}
	} else if resumeOrForkChoice == "resume" || resumeOrForkChoice == "fork" {
		// For resume or fork, get the Claude session ID
		if resumeOrForkChoice == "fork" {
			// For fork, we already have the new session ID in fm.choice
			claudeSessionID = fm.choice
		} else {
			// For resume, extract from session name
			claudeSessionID = extractClaudeSessionID(fm.sessionName)
			if claudeSessionID == "" {
				fmt.Fprintf(os.Stderr, "\n❌ Could not extract session ID for resume\n")
				os.Exit(1)
			}
		}

		// Update last used timestamp
		if err := updateLastUsed(fm.sessionPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
		}

		// Give terminal a moment to settle
		time.Sleep(100 * time.Millisecond)

		// Clear screen and show launching message
		fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
		fmt.Print("\033[0m")              // Reset all attributes

		if isFreshMemory {
			fmt.Printf("\n🔄 Launching fresh memory session\n")
		} else if resumeOrForkChoice == "fork" {
			fmt.Printf("\n✅ Launching forked session\n")
		} else {
			fmt.Printf("\n✅ Resuming Claude session\n")
		}
		fmt.Printf("📦 Session: %s\n", fm.sessionName)
		fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		if resumeOrForkChoice == "fork" {
			// For fork, start a new session with activation command
			relativeSessionPath := filepath.Join("sessions", filepath.Base(fm.sessionPath))
			activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", relativeSessionPath)

			launchCmd := exec.Command("claude", "--session-id", claudeSessionID, activationPrompt)
			launchCmd.Stdin = os.Stdin
			launchCmd.Stdout = os.Stdout
			launchCmd.Stderr = os.Stderr
			launchCmd.Env = os.Environ()

			if err := launchCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For resume, continue existing session
			resumeCmd := exec.Command("claude", "--resume", claudeSessionID)
			resumeCmd.Stdin = os.Stdin
			resumeCmd.Stdout = os.Stdout
			resumeCmd.Stderr = os.Stderr
			resumeCmd.Env = os.Environ()

			if err := resumeCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		// Load team-lead profile directly (skip profile selection menu)
		profileContent, err = loadComposedProfile("team-lead")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
			os.Exit(1)
		}
		profileName := "team-lead"

		// Give terminal a moment to settle
		time.Sleep(100 * time.Millisecond)

		// Clear screen and show launching message
		fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
		fmt.Print("\033[0m")              // Reset all attributes
		fmt.Printf("\n✅ Launching Claude with %s\n", profileName)
		fmt.Printf("📦 Session: %s\n", fm.sessionName)

		// Generate new Claude session ID
		claudeSessionID = uuid.New().String()

		// Rename session directory to include Claude session ID
		newSessionPath, err := renameSessionWithClaudeID(fm.sessionPath, fm.sessionName, claudeSessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error renaming session directory: %v\n", err)
			os.Exit(1)
		}

		// Update environment variable with new path
		if newSessionPath != "" {
			os.Setenv("CLAUDEX_SESSION_PATH", newSessionPath)
			fmt.Printf("📁 Session directory: %s\n", filepath.Base(newSessionPath))
			fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

			// Update last used timestamp
			if err := updateLastUsed(newSessionPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
			}
		}

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		// Construct relative session path for activation command
		relativeSessionPath := filepath.Join("sessions", filepath.Base(newSessionPath))
		activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", relativeSessionPath)

		// Launch the Claude session with activation command
		launchCmd := exec.Command("claude", "--session-id", claudeSessionID, activationPrompt)
		launchCmd.Stdin = os.Stdin
		launchCmd.Stdout = os.Stdout
		launchCmd.Stderr = os.Stderr
		launchCmd.Env = os.Environ() // Ensure environment variables are inherited

		if err := launchCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
			os.Exit(1)
		}
	}
}

func createNewSessionParallel(sessionsDir string, profileContent []byte) (string, string, string, error) {
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println()
	fmt.Println("\033[1;36m Create New Session \033[0m")
	fmt.Println()

	// Generate UUID for the session upfront
	claudeSessionID := uuid.New().String()

	// Get description from user
	fmt.Print("  Description: ")
	reader := bufio.NewReader(os.Stdin)
	description, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", err
	}
	description = strings.TrimSpace(description)

	if description == "" {
		return "", "", "", fmt.Errorf("description cannot be empty")
	}

	fmt.Println()
	fmt.Println("\033[90m  🤖 Generating session name...\033[0m")

	sessionName, err := generateSessionName(description)
	if err != nil {
		sessionName = createManualSlug(description)
	}

	// Create final session name with Claude session ID
	baseSessionName := sessionName
	sessionName = fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)

	// Ensure unique (in case of collision)
	originalName := sessionName
	counter := 1
	sessionPath := filepath.Join(sessionsDir, sessionName)
	for {
		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			break
		}
		sessionName = fmt.Sprintf("%s-%d", originalName, counter)
		sessionPath = filepath.Join(sessionsDir, sessionName)
		counter++
	}

	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return "", "", "", err
	}

	if err := os.WriteFile(filepath.Join(sessionPath, ".description"), []byte(description), 0644); err != nil {
		return "", "", "", err
	}

	created := time.Now().UTC().Format(time.RFC3339)
	if err := os.WriteFile(filepath.Join(sessionPath, ".created"), []byte(created), 0644); err != nil {
		return "", "", "", err
	}

	fmt.Println()
	fmt.Println("\033[1;32m  ✓ Created: " + sessionName + "\033[0m")
	fmt.Println()
	time.Sleep(500 * time.Millisecond)

	return sessionName, sessionPath, claudeSessionID, nil
}

func generateSessionName(description string) (string, error) {
	prompt := fmt.Sprintf("Generate a short, descriptive slug (2-4 words max, lowercase, hyphen-separated) for a work session based on this description: '%s'. Reply with ONLY the slug, nothing else. Examples: 'auth-refactor', 'api-performance-fix', 'user-dashboard-ui'", description)

	cmd := exec.Command("claude", "-p")
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`[a-z0-9-]+`)
	matches := re.FindAllString(string(output), -1)

	if len(matches) == 0 {
		return "", fmt.Errorf("no valid slug")
	}

	sessionName := matches[0]
	if len(sessionName) < 3 {
		return "", fmt.Errorf("slug too short")
	}

	return sessionName, nil
}

func createManualSlug(description string) string {
	slug := strings.ToLower(description)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if len(slug) > 50 {
		slug = slug[:50]
	}

	return slug
}

func getSessions(sessionsDir string) ([]sessionItem, error) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []sessionItem{}, nil
		}
		return nil, err
	}

	var sessions []sessionItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		var desc string
		var lastUsedTime time.Time
		var lastUsedStr string

		if data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name(), ".description")); err == nil {
			desc = strings.TrimSpace(string(data))
		}

		// Try to read last_used first, fall back to created
		if data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name(), ".last_used")); err == nil {
			lastUsedStr = strings.TrimSpace(string(data))
			if t, err := time.Parse(time.RFC3339, lastUsedStr); err == nil {
				lastUsedTime = t
				lastUsedStr = t.Format("2 Jan 2006 15:04:05")
			}
		} else if data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name(), ".created")); err == nil {
			// Fall back to created date if no last_used file
			lastUsedStr = strings.TrimSpace(string(data))
			if t, err := time.Parse(time.RFC3339, lastUsedStr); err == nil {
				lastUsedTime = t
				lastUsedStr = t.Format("2 Jan 2006 15:04:05")
			}
		}

		sessions = append(sessions, sessionItem{
			title:       entry.Name(),
			description: fmt.Sprintf("%s • %s", desc, lastUsedStr),
			created:     lastUsedTime,
			itemType:    "session",
		})
	}

	// Sort by last used date in descending order (most recently used first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].created.After(sessions[j].created)
	})

	return sessions, nil
}

func getProfiles() ([]string, error) {
	profileSet := make(map[string]bool)

	// Look for profiles in embedded FS profiles/agents/ directory
	entries, err := fs.ReadDir(profilesFS, "profiles/agents")
	if err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() && !strings.HasPrefix(name, ".") {
				profileSet[name] = true
			}
		}
	}

	// Also look for profiles in filesystem .claude/agents/ directory
	fsAgentsDir := filepath.Join(".claude", "agents")
	if fsEntries, err := os.ReadDir(fsAgentsDir); err == nil {
		for _, entry := range fsEntries {
			name := entry.Name()
			if !entry.IsDir() && !strings.HasPrefix(name, ".") {
				// Remove .md extension for consistent naming
				name = strings.TrimSuffix(name, ".md")
				profileSet[name] = true
			}
		}
	}

	// Convert set to sorted slice
	var profiles []string
	for name := range profileSet {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	return profiles, nil
}

func extractProfileDescription(profilePath string) string {
	file, err := profilesFS.Open(profilePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`(?i)(role:|principal|agent)`)

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			desc := strings.TrimLeft(line, "#*- ")
			desc = regexp.MustCompile(`(?i)role:`).ReplaceAllString(desc, "")
			desc = strings.TrimSpace(desc)
			if len(desc) > 60 {
				desc = desc[:60]
			}
			return desc
		}
	}

	return ""
}

func copyDir(src, dst string) error {
	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

func forkSession(sessionsDir, originalSessionName string) (string, string, string, error) {
	// Generate new UUID for the forked session
	claudeSessionID := uuid.New().String()

	// Strip the Claude session ID to get the base session name
	baseSessionName := stripClaudeSessionID(originalSessionName)

	// Also need to strip any existing fork counter (e.g., "my-task-2" -> "my-task")
	// Check if the last segment is a number
	lastHyphen := strings.LastIndex(baseSessionName, "-")
	if lastHyphen != -1 {
		potentialCounter := baseSessionName[lastHyphen+1:]
		// If it's just a number, strip it too
		if regexp.MustCompile(`^\d+$`).MatchString(potentialCounter) {
			baseSessionName = baseSessionName[:lastHyphen]
		}
	}

	// Create session name with new Claude session ID
	newSessionName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newSessionPath := filepath.Join(sessionsDir, newSessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	if err := copyDir(originalSessionPath, newSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

func freshMemorySession(sessionsDir, originalSessionName string) (string, string, string, error) {
	// Generate new UUID for the fresh session
	claudeSessionID := uuid.New().String()

	// Strip the Claude session ID to get the base session name
	baseSessionName := stripClaudeSessionID(originalSessionName)

	// Create session name with new Claude session ID (keep base slug)
	newSessionName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newSessionPath := filepath.Join(sessionsDir, newSessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	if err := copyDir(originalSessionPath, newSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	// DELETE the original folder (key difference from fork)
	if err := os.RemoveAll(originalSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to delete original session: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

func forkSessionWithDescription(sessionsDir, originalSessionName, description string) (string, string, string, error) {
	// Generate new UUID for the forked session
	claudeSessionID := uuid.New().String()

	// Generate new session name from description (like new session creation)
	baseSessionName, err := generateSessionName(description)
	if err != nil {
		// Fallback to manual slug if Claude API fails
		baseSessionName = createManualSlug(description)
	}

	// Create session name with new Claude session ID
	newSessionName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newSessionPath := filepath.Join(sessionsDir, newSessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	if err := copyDir(originalSessionPath, newSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	// Update .description file with new description
	descPath := filepath.Join(newSessionPath, ".description")
	if err := os.WriteFile(descPath, []byte(description), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write description: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

func hasClaudeSessionID(sessionName string) bool {
	// Claude session IDs are UUIDs in format: 8-4-4-4-12 hex digits
	// Example: 33342657-73dc-407d-9aa6-a28f2e619268
	uuidPattern := regexp.MustCompile(`-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidPattern.MatchString(sessionName)
}

func extractClaudeSessionID(sessionName string) string {
	if !hasClaudeSessionID(sessionName) {
		return ""
	}

	// Find the UUID pattern at the end
	uuidPattern := regexp.MustCompile(`-([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)
	matches := uuidPattern.FindStringSubmatch(sessionName)
	if len(matches) > 1 {
		return matches[1] // Return the captured UUID without the leading hyphen
	}
	return ""
}

func stripClaudeSessionID(sessionName string) string {
	// Claude session IDs are UUIDs in format: 8-4-4-4-12 hex digits
	// We want to strip the entire UUID, not just the last segment

	if !hasClaudeSessionID(sessionName) {
		return sessionName
	}

	// Remove the UUID pattern from the end
	uuidPattern := regexp.MustCompile(`-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidPattern.ReplaceAllString(sessionName, "")
}

func renameSessionWithClaudeID(oldPath, sessionName, claudeSessionID string) (string, error) {
	if oldPath == "" {
		// Ephemeral session, no directory to rename
		return "", nil
	}

	// Strip any existing Claude session ID from the session name
	baseSessionName := stripClaudeSessionID(sessionName)

	// Create new directory name with Claude session ID suffix
	parentDir := filepath.Dir(oldPath)
	newDirName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newPath := filepath.Join(parentDir, newDirName)

	// Rename the directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("failed to rename session directory: %w", err)
	}

	return newPath, nil
}

func loadProfile(profileName string) ([]byte, error) {
	// Look for profile in profiles/agents/ directory
	agentPath := "profiles/agents/" + profileName
	return fs.ReadFile(profilesFS, agentPath)
}

// loadProfileFromFS loads a profile from the filesystem (.claude/agents/)
func loadProfileFromFS(profileName string) ([]byte, error) {
	// Try with .md extension first
	agentPath := filepath.Join(".claude", "agents", profileName+".md")
	if data, err := os.ReadFile(agentPath); err == nil {
		return data, nil
	}

	// Try without extension
	agentPath = filepath.Join(".claude", "agents", profileName)
	return os.ReadFile(agentPath)
}

// loadComposedProfile tries embedded FS first, then filesystem
func loadComposedProfile(profileName string) ([]byte, error) {
	// First try embedded FS
	if data, err := loadProfile(profileName); err == nil {
		return data, nil
	}

	// Then try filesystem
	return loadProfileFromFS(profileName)
}

func resolveProfilePath(profileName string) string {
	// Look for profile in profiles/agents/ directory
	agentPath := "profiles/agents/" + profileName
	if _, err := fs.Stat(profilesFS, agentPath); err == nil {
		return agentPath
	}

	return ""
}

func updateLastUsed(sessionPath string) error {
	if sessionPath == "" {
		// Ephemeral session, no directory to update
		return nil
	}

	lastUsed := time.Now().UTC().Format(time.RFC3339)
	return os.WriteFile(filepath.Join(sessionPath, ".last_used"), []byte(lastUsed), 0644)
}
