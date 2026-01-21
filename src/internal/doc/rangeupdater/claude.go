package rangeupdater

import (
	"fmt"
	"log"
	"os/exec"

	"claudex/internal/services/commander"
	"claudex/internal/services/env"
	"claudex/internal/services/processregistry"
)

// InvokeClaudeForIndex invokes Claude to regenerate an index.md file.
// Claude uses its Edit tool to update the file directly.
// The recursion guard (CLAUDE_HOOK_INTERNAL=1) prevents infinite loops.
func InvokeClaudeForIndex(cmdr commander.Commander, env env.Environment, indexPath, listing, modifiedFiles string) error {
	// Recursion guard: check if we're already inside a hook invocation
	if env.Get("CLAUDE_HOOK_INTERNAL") == "1" {
		log.Printf("Skipping index update for %s: recursion guard triggered", indexPath)
		return nil
	}

	log.Printf("Spawning background process to regenerate %s", indexPath)

	// Build Claude prompt with context
	prompt := buildPrompt(indexPath, listing, modifiedFiles)

	// Create a detached background process using bash
	// This ensures the process survives even after the calling process exits
	// Claude will use its Edit tool to update the file directly
	// Using --model haiku for cost efficiency (index updates are simple tasks)
	bashScript := fmt.Sprintf(`
export CLAUDE_HOOK_INTERNAL=1
claude -p %q --model haiku 2>/dev/null
`, prompt)

	cmd := exec.Command("bash", "-c", bashScript)

	// Start the background process
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start background Claude process for %s: %v", indexPath, err)
		return fmt.Errorf("failed to start background Claude process: %w", err)
	}

	pid := cmd.Process.Pid
	log.Printf("Background process started (PID: %d) for %s", pid, indexPath)

	// Register PID for runaway process protection tracking
	processregistry.DefaultRegistry.Register(pid)

	// Launch goroutine to wait for process completion and clean up
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Background Claude process (PID: %d) exited with error: %v", pid, err)
		} else {
			log.Printf("Background Claude process (PID: %d) completed successfully", pid)
		}
		// Unregister PID when process exits
		processregistry.DefaultRegistry.Unregister(pid)
	}()

	return nil
}

// buildPrompt constructs the Claude prompt for index.md regeneration
func buildPrompt(indexPath, listing, modifiedFiles string) string {
	return fmt.Sprintf(`A code change was made. Update the index.md at %s if needed.

MODIFIED FILES:
%s

FILES IN DIRECTORY:
%s

This is a lightweight documentation pointer that helps developers understand the codebase. Other index.md files may exist in parent or child directories - explore them to understand the documentation structure and determine the appropriate scope for this file. Make thoughtful updates that keep it relevant and useful.`, indexPath, modifiedFiles, listing)
}
