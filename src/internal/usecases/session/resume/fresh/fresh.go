// Package fresh provides the use case for creating fresh memory sessions.
// It orchestrates copying session directories, clearing memory-related files,
// and deleting the original session to create a clean slate.
package fresh

import (
	"fmt"
	"path/filepath"

	"claudex/internal/services/filesystem"
	"claudex/internal/services/session"
	"claudex/internal/services/uuid"

	"github.com/spf13/afero"
)

// UseCase handles creating fresh memory sessions from existing sessions
type UseCase struct {
	fs          afero.Fs
	uuidGen     uuid.UUIDGenerator
	sessionsDir string
}

// New creates a new fresh memory use case
func New(fs afero.Fs, uuidGen uuid.UUIDGenerator, sessionsDir string) *UseCase {
	return &UseCase{
		fs:          fs,
		uuidGen:     uuidGen,
		sessionsDir: sessionsDir,
	}
}

// Execute creates a fresh memory session from an existing session by:
// 1. Generating a new UUID for the fresh session
// 2. Stripping the Claude session ID from the original session name to get the base name
// 3. Copying the session directory
// 4. Removing tracking files (.last-processed-line, etc.)
// 5. Resetting the doc update counter
// 6. Deleting the original session directory
// 7. Returning the new session info
func (uc *UseCase) Execute(originalSessionName string) (sessionName, sessionPath, claudeSessionID string, err error) {
	// Generate new UUID for the fresh session
	claudeSessionID = uc.uuidGen.New()

	// Strip the Claude session ID to get the base session name
	baseSessionName := session.StripClaudeSessionID(originalSessionName)

	// Create session name with new Claude session ID (keep base slug)
	sessionName = fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	sessionPath = filepath.Join(uc.sessionsDir, sessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(uc.sessionsDir, originalSessionName)
	if err := filesystem.CopyDir(uc.fs, originalSessionPath, sessionPath, false); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	// Reset tracking files for fresh session (new transcript starts at line 1)
	trackingFiles := []string{
		filepath.Join(sessionPath, ".last-processed-line-overview"),
		filepath.Join(sessionPath, ".last-processed-line"),
	}
	for _, f := range trackingFiles {
		uc.fs.Remove(f) // Ignore errors - file may not exist
	}

	// Reset doc update counter
	counterFile := filepath.Join(sessionPath, ".doc-update-counter")
	afero.WriteFile(uc.fs, counterFile, []byte("0"), 0644)

	// DELETE the original folder (key difference from fork)
	if err := uc.fs.RemoveAll(originalSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to delete original session: %w", err)
	}

	return sessionName, sessionPath, claudeSessionID, nil
}
