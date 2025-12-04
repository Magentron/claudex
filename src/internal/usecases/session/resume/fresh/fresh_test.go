package fresh

import (
	"path/filepath"
	"testing"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/require"
)

// Test_Execute_CopiesAndDeletesOriginal tests fresh memory session workflow
func Test_Execute_CopiesAndDeletesOriginal(t *testing.T) {
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

	// Create usecase and exercise
	uc := New(h.FS, h, sessionsDir)
	newSessionName, newSessionPath, claudeSessionID, err := uc.Execute(originalSessionName)

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
