package session

import (
	"fmt"
	"path/filepath"
	"strings"

	"claudex/internal/services/env"

	"github.com/spf13/afero"
)

// FindSessionFolder locates the session folder by ID using a priority-based search strategy.
// Priority 1: CLAUDEX_SESSION_PATH environment variable
// Priority 2: Pattern match in ./sessions/*-{sessionID}
// Returns the absolute path to the session folder or an error if not found.
func FindSessionFolder(fs afero.Fs, environment env.Environment, sessionID string) (string, error) {
	// Priority 1: Check environment variable
	if envPath := environment.Get("CLAUDEX_SESSION_PATH"); envPath != "" {
		// Verify the path exists
		exists, err := afero.DirExists(fs, envPath)
		if err != nil {
			return "", fmt.Errorf("failed to check CLAUDEX_SESSION_PATH: %w", err)
		}
		if exists {
			return envPath, nil
		}
		// If env var is set but path doesn't exist, that's an error
		return "", fmt.Errorf("CLAUDEX_SESSION_PATH is set but directory does not exist: %s", envPath)
	}

	// Priority 2: Pattern match in ./sessions/*-{sessionID}
	pattern := filepath.Join(".", "sessions", fmt.Sprintf("*-%s", sessionID))
	matches, err := afero.Glob(fs, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to glob session pattern: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("session folder not found for session ID: %s", sessionID)
	}

	// Return the first match (should be only one in practice)
	return matches[0], nil
}

// FindSessionFolderWithCwd is a variant that searches relative to a specific working directory.
// This is useful when the cwd is not the project root.
func FindSessionFolderWithCwd(fs afero.Fs, environment env.Environment, sessionID string, cwd string) (string, error) {
	// Priority 1: Check environment variable (absolute path)
	if envPath := environment.Get("CLAUDEX_SESSION_PATH"); envPath != "" {
		exists, err := afero.DirExists(fs, envPath)
		if err != nil {
			return "", fmt.Errorf("failed to check CLAUDEX_SESSION_PATH: %w", err)
		}
		if exists {
			return envPath, nil
		}
		return "", fmt.Errorf("CLAUDEX_SESSION_PATH is set but directory does not exist: %s", envPath)
	}

	// Priority 2: Pattern match in {cwd}/sessions/*-{sessionID}
	pattern := filepath.Join(cwd, "sessions", fmt.Sprintf("*-%s", sessionID))
	matches, err := afero.Glob(fs, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to glob session pattern: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("session folder not found for session ID: %s", sessionID)
	}

	return matches[0], nil
}

// GetSessionID extracts the session ID from a session folder path.
// It expects paths in the format: .../session-name-{uuid}
// Returns the UUID portion or empty string if not found.
func GetSessionID(sessionPath string) string {
	base := filepath.Base(sessionPath)
	parts := strings.Split(base, "-")

	// UUID format: 8-4-4-4-12 hex digits
	// We need at least 5 parts after splitting on dashes for a valid UUID
	if len(parts) < 5 {
		return ""
	}

	// Try to find UUID pattern (last 5 parts)
	uuidParts := parts[len(parts)-5:]

	// Validate rough UUID structure
	if len(uuidParts[0]) == 8 && len(uuidParts[1]) == 4 &&
		len(uuidParts[2]) == 4 && len(uuidParts[3]) == 4 &&
		len(uuidParts[4]) == 12 {
		return strings.Join(uuidParts, "-")
	}

	return ""
}
