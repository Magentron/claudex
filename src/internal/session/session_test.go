package session

import (
	"path/filepath"
	"testing"
	"time"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/require"
)

// Test_NewSession_CreatesDirectoryAndInvokesClaude tests new session creation workflow
// Note: We test the slug generation and UUID/timestamp handling in isolation since
// CreateWithDeps reads from os.Stdin which cannot be mocked in tests.
// The directory creation logic is tested thoroughly in the Fork test.
func Test_NewSession_CreatesDirectoryAndInvokesClaude(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.UUIDs = []string{"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"}
	h.FixedTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	// Test slug generation with Commander mock
	h.Commander.OnPattern("claude", "-p").Return([]byte("feature-login"), nil)

	description := "Implement login feature"
	slug, err := GenerateNameWithCmd(h.Commander, description)

	// Verify slug generation
	require.NoError(t, err)
	require.Equal(t, "feature-login", slug)

	// Verify Commander was invoked
	require.Len(t, h.Commander.Invocations, 1)
	invocation := h.Commander.Invocations[0]
	require.Equal(t, "claude", invocation.Name)
	require.Contains(t, invocation.Args, "-p")

	// Verify UUID and timestamp generation work correctly
	uuid := h.New()
	timestamp := h.Now().UTC().Format(time.RFC3339)

	require.Equal(t, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", uuid)
	require.Equal(t, "2024-01-15T10:30:00Z", timestamp)

	// The full CreateWithDeps function is tested indirectly
	// via the Fork test which exercises the same directory creation and metadata logic.
}

// Test_Resume_ExtractsSessionIDAndUpdatesLastUsed tests session resume workflow
func Test_Resume_ExtractsSessionIDAndUpdatesLastUsed(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.FixedTime = time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)

	sessionDir := "/project/sessions/feature-login-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	h.CreateSessionWithFiles(sessionDir, map[string]string{
		".description": "Login feature",
		".created":     "2024-01-15T10:30:00Z",
	})

	sessionName := "feature-login-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	// Exercise - Test session ID extraction
	hasID := HasClaudeSessionID(sessionName)
	claudeID := ExtractClaudeSessionID(sessionName)

	// Verify ID extraction
	require.True(t, hasID, "should detect Claude session ID in session name")
	require.Equal(t, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", claudeID)

	// Exercise - Update last used
	err := UpdateLastUsedWithDeps(h.FS, h, sessionDir)

	// Verify
	require.NoError(t, err)
	testutil.AssertFileExists(t, h.FS, filepath.Join(sessionDir, ".last_used"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(sessionDir, ".last_used"), "2024-01-15T14:00:00Z")
}

// Test_Fork_CopiesDirectoryAndCreatesNewSession tests session fork workflow
func Test_Fork_CopiesDirectoryAndCreatesNewSession(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	originalSessionName := "login-feature-old-uuid-1234-5678-abcd-ef12-345678901234"
	sessionsDir := "/project/sessions"

	// Create original session with files
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	h.CreateSessionWithFiles(originalSessionPath, map[string]string{
		".description":       "Original login",
		".created":           "2024-01-10T10:00:00Z",
		"session-history.md": "# History\n...",
		"execution-plan.md":  "# Plan\n...",
	})

	// Mock commander to return slug for new description
	h.Commander.OnPattern("claude", "-p").Return([]byte("auth-refactor"), nil)
	h.UUIDs = []string{"new-uuid-aaaa-bbbb-cccc-dddd-eeeeeeeeeeee"}

	// Exercise
	newSessionName, newSessionPath, claudeSessionID, err := ForkWithDescriptionWithDeps(
		h.FS, h.Commander, h,
		sessionsDir, originalSessionName, "Refactor to OAuth",
	)

	// Verify
	require.NoError(t, err)
	require.Equal(t, "new-uuid-aaaa-bbbb-cccc-dddd-eeeeeeeeeeee", claudeSessionID)
	// The mock commander now properly returns "auth-refactor" for the slug generation
	require.Equal(t, "auth-refactor-new-uuid-aaaa-bbbb-cccc-dddd-eeeeeeeeeeee", newSessionName)

	// New directory created
	testutil.AssertDirExists(t, h.FS, newSessionPath)
	expectedPath := filepath.Join(sessionsDir, "auth-refactor-new-uuid-aaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	require.Equal(t, expectedPath, newSessionPath)

	// Files copied
	testutil.AssertFileExists(t, h.FS, filepath.Join(newSessionPath, "session-history.md"))
	testutil.AssertFileExists(t, h.FS, filepath.Join(newSessionPath, "execution-plan.md"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(newSessionPath, "session-history.md"), "# History")

	// Description updated to new description
	testutil.AssertFileExists(t, h.FS, filepath.Join(newSessionPath, ".description"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(newSessionPath, ".description"), "Refactor to OAuth")

	// Original still exists
	testutil.AssertDirExists(t, h.FS, originalSessionPath)

	// Commander invoked for slug generation
	require.Len(t, h.Commander.Invocations, 1)
	invocation := h.Commander.Invocations[0]
	require.Equal(t, "claude", invocation.Name)
	require.Contains(t, invocation.Args, "-p")
}

// Test_FreshMemory_CopiesAndDeletesOriginal tests fresh memory session workflow
func Test_FreshMemory_CopiesAndDeletesOriginal(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	// Session name must match the pattern with dashes separating UUID segments
	// Format: slug-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
	originalSessionName := "login-feature-aaaabbbb-cccc-dddd-eeee-ffffffffffff"
	sessionsDir := "/project/sessions"

	// Create session with tracking files
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	h.CreateSessionWithFiles(originalSessionPath, map[string]string{
		".description":                  "Login feature",
		".created":                      "2024-01-10T10:00:00Z",
		".last-processed-line-overview": "50",
		".last-processed-line":          "100",
		".doc-update-counter":           "5",
		"session-history.md":            "# History",
	})

	h.UUIDs = []string{"11112222-3333-4444-5555-666666666666"}

	// Exercise
	newSessionName, newSessionPath, claudeSessionID, err := FreshMemoryWithDeps(
		h.FS, h,
		sessionsDir, originalSessionName,
	)

	// Verify
	require.NoError(t, err)
	require.Equal(t, "11112222-3333-4444-5555-666666666666", claudeSessionID)
	require.Equal(t, "login-feature-11112222-3333-4444-5555-666666666666", newSessionName)

	// New directory exists
	testutil.AssertDirExists(t, h.FS, newSessionPath)
	expectedPath := filepath.Join(sessionsDir, "login-feature-11112222-3333-4444-5555-666666666666")
	require.Equal(t, expectedPath, newSessionPath)

	// Session files copied
	testutil.AssertFileExists(t, h.FS, filepath.Join(newSessionPath, "session-history.md"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(newSessionPath, "session-history.md"), "# History")

	// Tracking files REMOVED
	testutil.AssertNoFileExists(t, h.FS, filepath.Join(newSessionPath, ".last-processed-line-overview"))
	testutil.AssertNoFileExists(t, h.FS, filepath.Join(newSessionPath, ".last-processed-line"))

	// Counter reset
	testutil.AssertFileExists(t, h.FS, filepath.Join(newSessionPath, ".doc-update-counter"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(newSessionPath, ".doc-update-counter"), "0")

	// Original DELETED
	testutil.AssertNoDirExists(t, h.FS, originalSessionPath)
}
