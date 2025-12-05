package fork

import (
	"path/filepath"
	"testing"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/require"
)

// Test_Execute_CopiesDirectoryAndCreatesNewSession tests session fork workflow
func Test_Execute_CopiesDirectoryAndCreatesNewSession(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	originalSessionName := "login-feature-12345678-abcd-ef12-3456-7890abcdef12"
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

	// Create usecase and exercise
	uc := New(h.FS, h.Commander, h, sessionsDir)
	newSessionName, newSessionPath, claudeSessionID, err := uc.Execute(
		originalSessionName, "Refactor to OAuth",
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
