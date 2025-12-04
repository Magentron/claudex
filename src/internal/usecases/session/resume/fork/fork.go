// Package fork provides the use case for forking existing sessions.
// It orchestrates copying session directories, generating new names and UUIDs,
// and updating metadata for the forked session.
package fork

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"claudex/internal/services/commander"
	"claudex/internal/services/filesystem"
	"claudex/internal/services/session"
	"claudex/internal/services/uuid"

	"github.com/spf13/afero"
)

// UseCase handles forking of existing sessions
type UseCase struct {
	fs          afero.Fs
	cmd         commander.Commander
	uuidGen     uuid.UUIDGenerator
	sessionsDir string
}

// New creates a new fork use case
func New(fs afero.Fs, cmd commander.Commander, uuidGen uuid.UUIDGenerator, sessionsDir string) *UseCase {
	return &UseCase{
		fs:          fs,
		cmd:         cmd,
		uuidGen:     uuidGen,
		sessionsDir: sessionsDir,
	}
}

// Execute forks an existing session by:
// 1. Generating a new UUID for the forked session
// 2. Stripping the Claude session ID from the original session name
// 3. Copying the session directory
// 4. Returning the new session info
func (uc *UseCase) Execute(originalSessionName string) (sessionName, sessionPath, claudeSessionID string, err error) {
	// Generate new UUID for the forked session
	claudeSessionID = uc.uuidGen.New()

	// Strip the Claude session ID to get the base session name
	baseSessionName := session.StripClaudeSessionID(originalSessionName)

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
	sessionName = fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	sessionPath = filepath.Join(uc.sessionsDir, sessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(uc.sessionsDir, originalSessionName)
	if err := filesystem.CopyDir(uc.fs, originalSessionPath, sessionPath, false); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	return sessionName, sessionPath, claudeSessionID, nil
}

// ExecuteWithDescription forks a session with a new description by:
// 1. Generating a new UUID for the forked session
// 2. Generating a new session name from the description (via Claude CLI or manual slug)
// 3. Copying the session directory
// 4. Updating the .description file with the new description
// 5. Returning the new session info
func (uc *UseCase) ExecuteWithDescription(originalSessionName, description string) (sessionName, sessionPath, claudeSessionID string, err error) {
	// Generate new UUID for the forked session
	claudeSessionID = uc.uuidGen.New()

	// Generate new session name from description (like new session creation)
	baseSessionName, err := session.GenerateNameWithCmd(uc.cmd, description)
	if err != nil {
		// Fallback to manual slug if Claude API fails
		baseSessionName = session.CreateManualSlug(description)
	}

	// Create session name with new Claude session ID
	sessionName = fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	sessionPath = filepath.Join(uc.sessionsDir, sessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(uc.sessionsDir, originalSessionName)
	if err := filesystem.CopyDir(uc.fs, originalSessionPath, sessionPath, false); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	// Update .description file with new description
	descPath := filepath.Join(sessionPath, ".description")
	if err := afero.WriteFile(uc.fs, descPath, []byte(description), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write Description: %w", err)
	}

	return sessionName, sessionPath, claudeSessionID, nil
}
