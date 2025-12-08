package session

import (
	"testing"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/require"
)

// Test_FindSessionFolder_EnvVarPriority tests that CLAUDEX_SESSION_PATH has priority
func Test_FindSessionFolder_EnvVarPriority(t *testing.T) {
	h := testutil.NewTestHarness()

	// Setup: Create session folder at env var path
	envPath := "/custom/session/path"
	h.CreateDir(envPath)

	// Also create a pattern-matching session (should be ignored)
	sessionID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	h.CreateDir("./sessions/feature-login-" + sessionID)

	// Set env var
	h.Env.Set("CLAUDEX_SESSION_PATH", envPath)

	// Exercise
	result, err := FindSessionFolder(h.FS, h.Env, sessionID)

	// Verify - env var takes priority
	require.NoError(t, err)
	require.Equal(t, envPath, result)
}

// Test_FindSessionFolder_EnvVarNotExists tests error when env var points to non-existent path
func Test_FindSessionFolder_EnvVarNotExists(t *testing.T) {
	h := testutil.NewTestHarness()

	// Set env var to non-existent path
	h.Env.Set("CLAUDEX_SESSION_PATH", "/does/not/exist")

	// Exercise
	sessionID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	_, err := FindSessionFolder(h.FS, h.Env, sessionID)

	// Verify - should error
	require.Error(t, err)
	require.Contains(t, err.Error(), "CLAUDEX_SESSION_PATH")
	require.Contains(t, err.Error(), "does not exist")
}

// Test_FindSessionFolder_PatternMatch tests finding session via glob pattern
func Test_FindSessionFolder_PatternMatch(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	sessionPath := "./sessions/feature-login-" + sessionID
	h.CreateDir(sessionPath)

	// Exercise
	result, err := FindSessionFolder(h.FS, h.Env, sessionID)

	// Verify - path should contain the session ID (afero may normalize path)
	require.NoError(t, err)
	require.Contains(t, result, sessionID)
	require.Contains(t, result, "sessions/feature-login")
}

// Test_FindSessionFolder_MultipleMatches tests behavior with multiple matching sessions
func Test_FindSessionFolder_MultipleMatches(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

	// Create multiple sessions with same ID (unlikely but possible)
	h.CreateDir("./sessions/feature-a-" + sessionID)
	h.CreateDir("./sessions/feature-b-" + sessionID)

	// Exercise
	result, err := FindSessionFolder(h.FS, h.Env, sessionID)

	// Verify - should return first match
	require.NoError(t, err)
	require.Contains(t, result, sessionID)
}

// Test_FindSessionFolder_NotFound tests error when no session is found
func Test_FindSessionFolder_NotFound(t *testing.T) {
	h := testutil.NewTestHarness()

	// Create sessions directory but no matching session
	h.CreateDir("./sessions")

	// Exercise
	sessionID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	_, err := FindSessionFolder(h.FS, h.Env, sessionID)

	// Verify
	require.Error(t, err)
	require.Contains(t, err.Error(), "session folder not found")
}

// Test_FindSessionFolderWithCwd_CustomCwd tests finding session with custom working directory
func Test_FindSessionFolderWithCwd_CustomCwd(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cwd := "/project/workspace"
	sessionPath := cwd + "/sessions/feature-login-" + sessionID
	h.CreateDir(sessionPath)

	// Exercise
	result, err := FindSessionFolderWithCwd(h.FS, h.Env, sessionID, cwd)

	// Verify
	require.NoError(t, err)
	require.Equal(t, sessionPath, result)
}

// Test_FindSessionFolderWithCwd_EnvVarOverridesCwd tests that env var still has priority with custom cwd
func Test_FindSessionFolderWithCwd_EnvVarOverridesCwd(t *testing.T) {
	h := testutil.NewTestHarness()

	// Setup env var path
	envPath := "/custom/session/path"
	h.CreateDir(envPath)
	h.Env.Set("CLAUDEX_SESSION_PATH", envPath)

	// Also create pattern-matching session in cwd
	sessionID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cwd := "/project/workspace"
	h.CreateDir(cwd + "/sessions/feature-login-" + sessionID)

	// Exercise
	result, err := FindSessionFolderWithCwd(h.FS, h.Env, sessionID, cwd)

	// Verify - env var takes priority
	require.NoError(t, err)
	require.Equal(t, envPath, result)
}

// Test_GetSessionID_ValidUUID tests extracting session ID from path
func Test_GetSessionID_ValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Standard session path",
			path:     "/sessions/feature-login-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			expected: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		},
		{
			name:     "Complex session name with dashes",
			path:     "./sessions/my-feature-work-11112222-3333-4444-5555-666666666666",
			expected: "11112222-3333-4444-5555-666666666666",
		},
		{
			name:     "Session name without UUID",
			path:     "/sessions/feature-login",
			expected: "",
		},
		{
			name:     "Malformed UUID",
			path:     "/sessions/feature-login-invalid-uuid-format",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSessionID(tt.path)
			require.Equal(t, tt.expected, result)
		})
	}
}
