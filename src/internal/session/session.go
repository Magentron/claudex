// Package session provides session management functionality for Claudex.
// It handles creating, forking, resuming, and managing Claude CLI sessions
// with support for session naming, metadata tracking, and filesystem operations.
package session

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"claudex/internal/fsutil"
	"claudex/internal/ui"

	"github.com/spf13/afero"
)

// Commander abstracts process execution for testability
type Commander interface {
	// Run executes command and returns combined output
	Run(name string, args ...string) ([]byte, error)
	// Start launches interactive command with stdio attached
	Start(name string, stdin io.Reader, stdout, stderr io.Writer, args ...string) error
}

// Clock abstracts time for testability
type Clock interface {
	Now() time.Time
}

// UUIDGenerator abstracts UUID generation for testability
type UUIDGenerator interface {
	New() string
}

// CreateWithDeps creates a new session using injected dependencies
func CreateWithDeps(fs afero.Fs, cmd Commander, uuidGen UUIDGenerator, clock Clock, sessionsDir string, profileContent []byte) (string, string, string, error) {
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println()
	fmt.Println("\033[1;36m Create New Session \033[0m")
	fmt.Println()

	// Generate UUID for the session upfront
	claudeSessionID := uuidGen.New()

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

	sessionName, err := GenerateNameWithCmd(cmd, description)
	if err != nil {
		sessionName = CreateManualSlug(description)
	}

	// Create final session name with Claude session ID
	baseSessionName := sessionName
	sessionName = fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)

	// Ensure unique (in case of collision)
	originalName := sessionName
	counter := 1
	sessionPath := filepath.Join(sessionsDir, sessionName)
	for {
		if _, err := fs.Stat(sessionPath); os.IsNotExist(err) {
			break
		}
		sessionName = fmt.Sprintf("%s-%d", originalName, counter)
		sessionPath = filepath.Join(sessionsDir, sessionName)
		counter++
	}

	if err := fs.MkdirAll(sessionPath, 0755); err != nil {
		return "", "", "", err
	}

	if err := afero.WriteFile(fs, filepath.Join(sessionPath, ".description"), []byte(description), 0644); err != nil {
		return "", "", "", err
	}

	created := clock.Now().UTC().Format(time.RFC3339)
	if err := afero.WriteFile(fs, filepath.Join(sessionPath, ".created"), []byte(created), 0644); err != nil {
		return "", "", "", err
	}

	fmt.Println()
	fmt.Println("\033[1;32m  ✓ Created: " + sessionName + "\033[0m")
	fmt.Println()
	time.Sleep(500 * time.Millisecond)

	return sessionName, sessionPath, claudeSessionID, nil
}

// Create is a wrapper that uses default dependencies from main package
// Note: This should not be used directly in production code; use CreateWithDeps instead
func Create(fs afero.Fs, cmd Commander, uuidGen UUIDGenerator, clock Clock, sessionsDir string, profileContent []byte) (string, string, string, error) {
	return CreateWithDeps(fs, cmd, uuidGen, clock, sessionsDir, profileContent)
}

// GetSessions retrieves all sessions from the sessions directory
func GetSessions(fs afero.Fs, sessionsDir string) ([]ui.SessionItem, error) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ui.SessionItem{}, nil
		}
		return nil, err
	}

	var sessions []ui.SessionItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		var desc string
		var lastUsedTime time.Time
		var lastUsedStr string

		if data, err := afero.ReadFile(fs, filepath.Join(sessionsDir, entry.Name(), ".description")); err == nil {
			desc = strings.TrimSpace(string(data))
		}

		// Try to read last_used first, fall back to created
		if data, err := afero.ReadFile(fs, filepath.Join(sessionsDir, entry.Name(), ".last_used")); err == nil {
			lastUsedStr = strings.TrimSpace(string(data))
			if t, err := time.Parse(time.RFC3339, lastUsedStr); err == nil {
				lastUsedTime = t
				lastUsedStr = t.Format("2 Jan 2006 15:04:05")
			}
		} else if data, err := afero.ReadFile(fs, filepath.Join(sessionsDir, entry.Name(), ".created")); err == nil {
			// Fall back to created date if no last_used file
			lastUsedStr = strings.TrimSpace(string(data))
			if t, err := time.Parse(time.RFC3339, lastUsedStr); err == nil {
				lastUsedTime = t
				lastUsedStr = t.Format("2 Jan 2006 15:04:05")
			}
		}

		sessions = append(sessions, ui.SessionItem{
			Title:       entry.Name(),
			Description: fmt.Sprintf("%s • %s", desc, lastUsedStr),
			Created:     lastUsedTime,
			ItemType:    "session",
		})
	}

	// Sort by last used date in descending order (most recently used first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Created.After(sessions[j].Created)
	})

	return sessions, nil
}

// Fork creates a new session by copying an existing session
func Fork(fs afero.Fs, uuidGen UUIDGenerator, sessionsDir, originalSessionName string) (string, string, string, error) {
	// Generate new UUID for the forked session
	claudeSessionID := uuidGen.New()

	// Strip the Claude session ID to get the base session name
	baseSessionName := StripClaudeSessionID(originalSessionName)

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
	if err := fsutil.CopyDir(fs, originalSessionPath, newSessionPath, false); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

// FreshMemoryWithDeps creates a fresh session with cleared memory using injected dependencies
func FreshMemoryWithDeps(fs afero.Fs, uuidGen UUIDGenerator, sessionsDir, originalSessionName string) (string, string, string, error) {
	// Generate new UUID for the fresh session
	claudeSessionID := uuidGen.New()

	// Strip the Claude session ID to get the base session name
	baseSessionName := StripClaudeSessionID(originalSessionName)

	// Create session name with new Claude session ID (keep base slug)
	newSessionName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newSessionPath := filepath.Join(sessionsDir, newSessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	if err := fsutil.CopyDir(fs, originalSessionPath, newSessionPath, false); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	// Reset tracking files for fresh session (new transcript starts at line 1)
	trackingFiles := []string{
		filepath.Join(newSessionPath, ".last-processed-line-overview"),
		filepath.Join(newSessionPath, ".last-processed-line"),
	}
	for _, f := range trackingFiles {
		fs.Remove(f) // Ignore errors - file may not exist
	}

	// Reset doc update counter
	counterFile := filepath.Join(newSessionPath, ".doc-update-counter")
	afero.WriteFile(fs, counterFile, []byte("0"), 0644)

	// DELETE the original folder (key difference from fork)
	if err := fs.RemoveAll(originalSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to delete original session: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

// FreshMemory is a wrapper that uses default dependencies
// Note: This should not be used directly in production code; use FreshMemoryWithDeps instead
func FreshMemory(fs afero.Fs, uuidGen UUIDGenerator, sessionsDir, originalSessionName string) (string, string, string, error) {
	return FreshMemoryWithDeps(fs, uuidGen, sessionsDir, originalSessionName)
}

// ForkWithDescriptionWithDeps forks a session with a new description using injected dependencies
func ForkWithDescriptionWithDeps(fs afero.Fs, cmd Commander, uuidGen UUIDGenerator, sessionsDir, originalSessionName, description string) (string, string, string, error) {
	// Generate new UUID for the forked session
	claudeSessionID := uuidGen.New()

	// Generate new session name from description (like new session creation)
	baseSessionName, err := GenerateNameWithCmd(cmd, description)
	if err != nil {
		// Fallback to manual slug if Claude API fails
		baseSessionName = CreateManualSlug(description)
	}

	// Create session name with new Claude session ID
	newSessionName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newSessionPath := filepath.Join(sessionsDir, newSessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	if err := fsutil.CopyDir(fs, originalSessionPath, newSessionPath, false); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	// Update .description file with new description
	descPath := filepath.Join(newSessionPath, ".description")
	if err := afero.WriteFile(fs, descPath, []byte(description), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write Description: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

// ForkWithDescription is a wrapper that uses default dependencies
// Note: This should not be used directly in production code; use ForkWithDescriptionWithDeps instead
func ForkWithDescription(fs afero.Fs, cmd Commander, uuidGen UUIDGenerator, sessionsDir, originalSessionName, description string) (string, string, string, error) {
	return ForkWithDescriptionWithDeps(fs, cmd, uuidGen, sessionsDir, originalSessionName, description)
}

// UpdateLastUsedWithDeps updates the last used timestamp using injected dependencies
func UpdateLastUsedWithDeps(fs afero.Fs, clock Clock, sessionPath string) error {
	if sessionPath == "" {
		// Ephemeral session, no directory to update
		return nil
	}

	lastUsed := clock.Now().UTC().Format(time.RFC3339)
	return afero.WriteFile(fs, filepath.Join(sessionPath, ".last_used"), []byte(lastUsed), 0644)
}

// UpdateLastUsed is a wrapper that uses default dependencies
// Note: This should not be used directly in production code; use UpdateLastUsedWithDeps instead
func UpdateLastUsed(fs afero.Fs, clock Clock, sessionPath string) error {
	return UpdateLastUsedWithDeps(fs, clock, sessionPath)
}

// GenerateNameWithCmd generates a session name using the provided Commander
func GenerateNameWithCmd(commander Commander, description string) (string, error) {
	prompt := fmt.Sprintf("Generate a short, descriptive slug (2-4 words max, lowercase, hyphen-separated) for a work session based on this Description: '%s'. Reply with ONLY the slug, nothing else. Examples: 'auth-refactor', 'api-performance-fix', 'user-dashboard-ui'", description)

	// Create a pipe to capture output
	var stdout bytes.Buffer
	stdin := strings.NewReader(prompt)

	// Use Start method which supports stdin/stdout/stderr
	err := commander.Start("claude", stdin, &stdout, os.Stderr, "-p")
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

// GenerateName generates a session name using the default Commander
// Note: This should not be used directly in production code; use GenerateNameWithCmd instead
func GenerateName(commander Commander, description string) (string, error) {
	return GenerateNameWithCmd(commander, description)
}
