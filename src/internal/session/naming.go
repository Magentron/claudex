package session

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
func RenameWithClaudeID(oldPath, sessionName, claudeSessionID string) (string, error) {
	if oldPath == "" {
		// Ephemeral session, no directory to rename
		return "", nil
	}

	// Strip any existing Claude session ID from the session name
	baseSessionName := StripClaudeSessionID(sessionName)

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
