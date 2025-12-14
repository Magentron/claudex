package rangeupdater

import (
	"fmt"
	"log"
	"os/exec"

	"claudex/internal/services/commander"
	"claudex/internal/services/env"
)

// InvokeClaudeForIndex invokes Claude to regenerate an index.md file.
// This follows the pattern from indexupdater.go:153-199 with a detached background process.
// The recursion guard (CLAUDE_HOOK_INTERNAL=1) prevents infinite loops.
func InvokeClaudeForIndex(cmdr commander.Commander, env env.Environment, indexDir, indexPath, listing, modifiedFiles string) error {
	// Recursion guard: check if we're already inside a hook invocation
	if env.Get("CLAUDE_HOOK_INTERNAL") == "1" {
		log.Printf("Skipping index update for %s: recursion guard triggered", indexPath)
		return nil
	}

	log.Printf("Spawning background process to regenerate %s", indexPath)

	// Build Claude prompt with context
	prompt := buildPrompt(indexDir, indexPath, listing, modifiedFiles)

	// Create a detached background process using bash
	// This ensures the process survives even after the calling process exits
	// Note: Claude CLI outputs to stdout, so we pipe to the file
	// Using --model haiku for cost efficiency (index updates are simple tasks)
	bashScript := fmt.Sprintf(`
export CLAUDE_HOOK_INTERNAL=1
claude -p %q --allowedTools "" --model haiku > %q 2>/dev/null
`, prompt, indexPath)

	cmd := exec.Command("bash", "-c", bashScript)

	// Detach the process so it survives after we exit
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start background Claude process for %s: %v", indexPath, err)
		return fmt.Errorf("failed to start background Claude process: %w", err)
	}

	log.Printf("Background process started (PID: %d) for %s", cmd.Process.Pid, indexPath)
	return nil
}

// buildPrompt constructs the Claude prompt for index.md regeneration
func buildPrompt(indexDir, indexPath, listing, modifiedFiles string) string {
	return fmt.Sprintf(`You are regenerating an index.md file for a documentation directory.

DIRECTORY: %s
INDEX FILE: %s
MODIFIED FILES:
%s

FILES IN DIRECTORY:
%s

INSTRUCTIONS:
1. Read the existing index.md to understand the current structure and style
2. Generate updated index.md content listing all files in this directory
3. Use minimal pointer style: brief one-line descriptions
4. Group files logically if patterns exist
5. Keep descriptions concise (one line per file)
6. Output ONLY the markdown content for index.md (no explanations or metadata)

Read the existing index.md first, then generate the updated content.`, indexDir, indexPath, modifiedFiles, listing)
}
