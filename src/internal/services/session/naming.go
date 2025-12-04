// Package session provides session management services for Claudex.
// It handles session metadata, storage operations, and naming utilities.
package session

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"claudex/internal/services/commander"

	"github.com/spf13/afero"
)

// CreateManualSlug creates a slug from description as a fallback
func CreateManualSlug(description string) string {
	slug := strings.ToLower(description)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if len(slug) > 50 {
		slug = slug[:50]
	}

	return slug
}

// HasClaudeSessionID checks if a session name contains a Claude session ID
func HasClaudeSessionID(sessionName string) bool {
	// Claude session IDs are UUIDs in format: 8-4-4-4-12 hex digits
	// Example: 33342657-73dc-407d-9aa6-a28f2e619268
	uuidPattern := regexp.MustCompile(`-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidPattern.MatchString(sessionName)
}

// ExtractClaudeSessionID extracts the Claude session ID from a session name
func ExtractClaudeSessionID(sessionName string) string {
	if !HasClaudeSessionID(sessionName) {
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

// StripClaudeSessionID removes the Claude session ID from a session name
func StripClaudeSessionID(sessionName string) string {
	// Claude session IDs are UUIDs in format: 8-4-4-4-12 hex digits
	// We want to strip the entire UUID, not just the last segment

	if !HasClaudeSessionID(sessionName) {
		return sessionName
	}

	// Remove the UUID pattern from the end
	uuidPattern := regexp.MustCompile(`-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidPattern.ReplaceAllString(sessionName, "")
}

// RenameWithClaudeID renames a session directory to include the Claude session ID
func RenameWithClaudeID(fs afero.Fs, sessionPath, claudeSessionID string) error {
	if sessionPath == "" {
		// Ephemeral session, no directory to rename
		return nil
	}

	// Extract session name from path
	sessionName := filepath.Base(sessionPath)

	// Strip any existing Claude session ID from the session name
	baseSessionName := StripClaudeSessionID(sessionName)

	// Create new directory name with Claude session ID suffix
	parentDir := filepath.Dir(sessionPath)
	newDirName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newPath := filepath.Join(parentDir, newDirName)

	// Rename the directory
	if err := fs.Rename(sessionPath, newPath); err != nil {
		return fmt.Errorf("failed to rename session directory: %w", err)
	}

	return nil
}

// GenerateNameWithCmd generates a session name using the provided Commander
func GenerateNameWithCmd(cmd commander.Commander, description string) (string, error) {
	prompt := fmt.Sprintf("Generate a short, descriptive slug (2-4 words max, lowercase, hyphen-separated) for a work session based on this Description: '%s'. Reply with ONLY the slug, nothing else. Examples: 'auth-refactor', 'api-performance-fix', 'user-dashboard-ui'", description)

	// Create a pipe to capture output
	var stdout bytes.Buffer
	stdin := strings.NewReader(prompt)

	// Use Start method which supports stdin/stdout/stderr
	err := cmd.Start("claude", stdin, &stdout, os.Stderr, "-p")
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`[a-z0-9-]+`)
	matches := re.FindAllString(stdout.String(), -1)

	if len(matches) == 0 {
		return "", fmt.Errorf("no valid slug")
	}

	sessionName := matches[0]
	if len(sessionName) < 3 {
		return "", fmt.Errorf("slug too short")
	}

	return sessionName, nil
}
