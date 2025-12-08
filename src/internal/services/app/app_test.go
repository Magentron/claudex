package app

import (
	"path/filepath"
	"testing"

	"claudex/internal/testutil"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenameLogFileForSession_NewSession verifies log file rename for new non-ephemeral sessions
// Given: Timestamp-based log file exists, non-ephemeral session info
// When: renameLogFileForSession called
// Then: Log file renamed to {session-name}.log, env var updated
func TestRenameLogFileForSession_NewSession(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	projectDir := "/project"
	logsDir := filepath.Join(projectDir, "logs")
	h.CreateDir(logsDir)

	// Create initial timestamp-based log file
	timestampLogPath := filepath.Join(logsDir, "claudex-20241208-120000.log")
	h.WriteFile(timestampLogPath, "[claudex] Initial log entry\n")

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		logFilePath: timestampLogPath,
	}

	// Set initial env var
	h.Env.Set("CLAUDEX_LOG_FILE", timestampLogPath)

	// Create session info for new session
	si := SessionInfo{
		Name: "session-feature-x-abc123",
		Path: filepath.Join(projectDir, "sessions", "session-feature-x-abc123"),
		Mode: LaunchModeNew,
	}

	// Execute rename
	app.renameLogFileForSession(si)

	// Assert: Log file renamed
	expectedNewPath := filepath.Join(logsDir, "session-feature-x-abc123.log")
	exists, err := afero.Exists(h.FS, expectedNewPath)
	require.NoError(t, err)
	assert.True(t, exists, "renamed log file should exist at %s", expectedNewPath)

	// Assert: Original timestamp log no longer exists (renamed)
	exists, err = afero.Exists(h.FS, timestampLogPath)
	require.NoError(t, err)
	assert.False(t, exists, "original timestamp log should be renamed (not exist)")

	// Assert: App state updated
	assert.Equal(t, expectedNewPath, app.logFilePath, "app.logFilePath should be updated")

	// Assert: Environment variable updated
	assert.Equal(t, expectedNewPath, h.Env.Get("CLAUDEX_LOG_FILE"), "CLAUDEX_LOG_FILE env var should be updated")

	// Assert: Log content preserved
	content, err := afero.ReadFile(h.FS, expectedNewPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Initial log entry", "log content should be preserved after rename")
}

// TestRenameLogFileForSession_EphemeralSession verifies ephemeral sessions keep timestamp names
// Given: Timestamp-based log file exists, ephemeral session (empty path)
// When: renameLogFileForSession called
// Then: Log file NOT renamed, keeps timestamp name
func TestRenameLogFileForSession_EphemeralSession(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	projectDir := "/project"
	logsDir := filepath.Join(projectDir, "logs")
	h.CreateDir(logsDir)

	// Create initial timestamp-based log file
	timestampLogPath := filepath.Join(logsDir, "claudex-20241208-130000.log")
	h.WriteFile(timestampLogPath, "[claudex] Ephemeral session log\n")

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		logFilePath: timestampLogPath,
	}

	// Set initial env var
	h.Env.Set("CLAUDEX_LOG_FILE", timestampLogPath)

	// Create session info for ephemeral mode
	si := SessionInfo{
		Name: "ephemeral",
		Path: "", // Empty path indicates ephemeral
		Mode: LaunchModeEphemeral,
	}

	// Execute rename (should be a no-op)
	app.renameLogFileForSession(si)

	// Assert: Original timestamp log still exists
	exists, err := afero.Exists(h.FS, timestampLogPath)
	require.NoError(t, err)
	assert.True(t, exists, "timestamp log should NOT be renamed for ephemeral sessions")

	// Assert: No session-named log created
	sessionLogPath := filepath.Join(logsDir, "ephemeral.log")
	exists, err = afero.Exists(h.FS, sessionLogPath)
	require.NoError(t, err)
	assert.False(t, exists, "should not create session-named log for ephemeral")

	// Assert: App state unchanged
	assert.Equal(t, timestampLogPath, app.logFilePath, "app.logFilePath should remain unchanged")

	// Assert: Environment variable unchanged
	assert.Equal(t, timestampLogPath, h.Env.Get("CLAUDEX_LOG_FILE"), "CLAUDEX_LOG_FILE should remain unchanged for ephemeral")
}

// TestRenameLogFileForSession_ResumeWithExistingLog verifies resume appends to existing session log
// Given: Session log already exists from previous invocation
// When: renameLogFileForSession called
// Then: Appends to existing session log, timestamp log removed
func TestRenameLogFileForSession_ResumeWithExistingLog(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	projectDir := "/project"
	logsDir := filepath.Join(projectDir, "logs")
	h.CreateDir(logsDir)

	// Create existing session log from previous run
	sessionLogPath := filepath.Join(logsDir, "session-resume-test-xyz789.log")
	h.WriteFile(sessionLogPath, "[claudex] Previous session log entry\n")

	// Create new timestamp-based log file for this invocation
	timestampLogPath := filepath.Join(logsDir, "claudex-20241208-140000.log")
	h.WriteFile(timestampLogPath, "[claudex] New invocation log entry\n")

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		logFilePath: timestampLogPath,
	}

	// Set initial env var
	h.Env.Set("CLAUDEX_LOG_FILE", timestampLogPath)

	// Create session info for resume
	si := SessionInfo{
		Name: "session-resume-test-xyz789",
		Path: filepath.Join(projectDir, "sessions", "session-resume-test-xyz789"),
		Mode: LaunchModeResume,
	}

	// Execute rename
	app.renameLogFileForSession(si)

	// Assert: Session log exists
	exists, err := afero.Exists(h.FS, sessionLogPath)
	require.NoError(t, err)
	assert.True(t, exists, "session log should exist")

	// Assert: Original timestamp log no longer exists (moved/deleted)
	exists, err = afero.Exists(h.FS, timestampLogPath)
	require.NoError(t, err)
	assert.False(t, exists, "timestamp log should be removed after rename")

	// Assert: App state updated
	assert.Equal(t, sessionLogPath, app.logFilePath, "app.logFilePath should point to session log")

	// Assert: Environment variable updated
	assert.Equal(t, sessionLogPath, h.Env.Get("CLAUDEX_LOG_FILE"), "CLAUDEX_LOG_FILE should point to session log")

	// Assert: Session log contains previous content
	content, err := afero.ReadFile(h.FS, sessionLogPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Previous session log entry", "should preserve previous log content")
}

// TestRenameLogFileForSession_RenameFailureGraceful verifies graceful handling of rename failures
// Given: Rename would fail (e.g., permissions, disk full)
// When: renameLogFileForSession called
// Then: Warning logged, original log file still usable, app continues
func TestRenameLogFileForSession_RenameFailureGraceful(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	projectDir := "/project"
	logsDir := filepath.Join(projectDir, "logs")
	h.CreateDir(logsDir)

	// Create initial timestamp-based log file
	timestampLogPath := filepath.Join(logsDir, "claudex-20241208-150000.log")
	h.WriteFile(timestampLogPath, "[claudex] Initial log entry\n")

	// Use a read-only filesystem to simulate rename failure
	roFS := afero.NewReadOnlyFs(h.FS)

	// Create app with read-only filesystem
	app := &App{
		deps: &Dependencies{
			FS:    roFS, // Read-only will cause rename to fail
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		logFilePath: timestampLogPath,
	}

	// Set initial env var
	h.Env.Set("CLAUDEX_LOG_FILE", timestampLogPath)

	// Create session info
	si := SessionInfo{
		Name: "session-will-fail-rename",
		Path: filepath.Join(projectDir, "sessions", "session-will-fail-rename"),
		Mode: LaunchModeNew,
	}

	// Execute rename (should fail gracefully and not panic)
	assert.NotPanics(t, func() {
		app.renameLogFileForSession(si)
	}, "should not panic on failed rename")

	// Assert: Original timestamp log still exists (rename failed)
	exists, err := afero.Exists(h.FS, timestampLogPath)
	require.NoError(t, err)
	assert.True(t, exists, "original log should still exist after failed rename")

	// Assert: Environment variable should remain as original (or unchanged)
	// The implementation should handle this gracefully
	envVar := h.Env.Get("CLAUDEX_LOG_FILE")
	assert.NotEmpty(t, envVar, "CLAUDEX_LOG_FILE should still be set after failed rename")
}

// TestRenameLogFileForSession_EnvVarUpdated verifies CLAUDEX_LOG_FILE env var is updated
// Given: Successful rename
// Then: Environment Get("CLAUDEX_LOG_FILE") returns new path
func TestRenameLogFileForSession_EnvVarUpdated(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	projectDir := "/project"
	logsDir := filepath.Join(projectDir, "logs")
	h.CreateDir(logsDir)

	// Create initial timestamp-based log file
	timestampLogPath := filepath.Join(logsDir, "claudex-20241208-160000.log")
	h.WriteFile(timestampLogPath, "[claudex] Testing env var update\n")

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		logFilePath: timestampLogPath,
	}

	// Set initial env var
	h.Env.Set("CLAUDEX_LOG_FILE", timestampLogPath)
	initialEnvVar := h.Env.Get("CLAUDEX_LOG_FILE")
	assert.Equal(t, timestampLogPath, initialEnvVar, "initial env var should match timestamp log")

	// Create session info
	si := SessionInfo{
		Name: "session-env-test-def456",
		Path: filepath.Join(projectDir, "sessions", "session-env-test-def456"),
		Mode: LaunchModeNew,
	}

	// Execute rename
	app.renameLogFileForSession(si)

	// Assert: Environment variable updated to new path
	expectedNewPath := filepath.Join(logsDir, "session-env-test-def456.log")
	updatedEnvVar := h.Env.Get("CLAUDEX_LOG_FILE")
	assert.Equal(t, expectedNewPath, updatedEnvVar, "CLAUDEX_LOG_FILE should be updated to new session log path")
	assert.NotEqual(t, initialEnvVar, updatedEnvVar, "env var should have changed from initial value")
}

// TestRenameLogFileForSession_EmptyLogFilePath verifies handling of empty logFilePath
// Given: app.logFilePath is empty
// When: renameLogFileForSession called
// Then: No panic, graceful no-op
func TestRenameLogFileForSession_EmptyLogFilePath(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	projectDir := "/project"

	// Create app with empty logFilePath
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		logFilePath: "", // Empty path
	}

	// Create session info
	si := SessionInfo{
		Name: "session-empty-path",
		Path: filepath.Join(projectDir, "sessions", "session-empty-path"),
		Mode: LaunchModeNew,
	}

	// Execute rename (should not panic)
	assert.NotPanics(t, func() {
		app.renameLogFileForSession(si)
	}, "should not panic with empty logFilePath")
}

// TestRenameLogFileForSession_ForkMode verifies fork creates new session log
// Given: Fork mode session
// When: renameLogFileForSession called
// Then: New log file created with forked session name
func TestRenameLogFileForSession_ForkMode(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	projectDir := "/project"
	logsDir := filepath.Join(projectDir, "logs")
	h.CreateDir(logsDir)

	// Create initial timestamp-based log file
	timestampLogPath := filepath.Join(logsDir, "claudex-20241208-170000.log")
	h.WriteFile(timestampLogPath, "[claudex] Fork session log\n")

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		logFilePath: timestampLogPath,
	}

	// Set initial env var
	h.Env.Set("CLAUDEX_LOG_FILE", timestampLogPath)

	// Create session info for fork (new session name, but forked from original)
	si := SessionInfo{
		Name:         "session-forked-from-original-ghi789",
		Path:         filepath.Join(projectDir, "sessions", "session-forked-from-original-ghi789"),
		Mode:         LaunchModeFork,
		OriginalName: "session-original-abc123",
	}

	// Execute rename
	app.renameLogFileForSession(si)

	// Assert: New log file created with forked session name
	expectedNewPath := filepath.Join(logsDir, "session-forked-from-original-ghi789.log")
	exists, err := afero.Exists(h.FS, expectedNewPath)
	require.NoError(t, err)
	assert.True(t, exists, "forked session should have its own log file")

	// Assert: App state updated
	assert.Equal(t, expectedNewPath, app.logFilePath, "app.logFilePath should point to forked session log")

	// Assert: Environment variable updated
	assert.Equal(t, expectedNewPath, h.Env.Get("CLAUDEX_LOG_FILE"), "CLAUDEX_LOG_FILE should point to forked session log")
}
