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

// TestInit_CreatesClaudexDirectory verifies Init creates .claudex/ structure on fresh filesystem
// Given: Fresh filesystem (no .claudex/)
// When: Init() called
// Then: .claudex/ folder created with config.toml defaults
func TestInit_CreatesClaudexDirectory(t *testing.T) {
	// Setup - fresh filesystem
	h := testutil.NewTestHarness()

	// Create app with mocked dependencies
	showVersion := false
	noOverwrite := false
	updateDocs := false
	setupMCP := false
	createIndex := ""
	docPaths := []string{}

	app := &App{
		deps:            &Dependencies{FS: h.FS, Cmd: h.Commander, Clock: h, UUID: h, Env: h.Env},
		version:         "1.0.0",
		showVersion:     &showVersion,
		noOverwriteFlag: &noOverwrite,
		updateDocsFlag:  &updateDocs,
		setupMCPFlag:    &setupMCP,
		createIndexFlag: &createIndex,
		docPathsFlag:    docPaths,
	}

	// Mock environment variables
	h.Env.Set("HOME", "/home/user")

	// Execute Init (migration uses relative paths like .claudex, which work with MemMapFs)
	err := app.Init()
	require.NoError(t, err, "Init should succeed on fresh filesystem")

	// Assert: .claudex/ directory created (using paths.ClaudexDir constant)
	exists, err := afero.DirExists(h.FS, ".claudex")
	require.NoError(t, err)
	assert.True(t, exists, ".claudex/ directory should be created")

	// Assert: config.toml created with defaults
	exists, err = afero.Exists(h.FS, ".claudex/config.toml")
	require.NoError(t, err)
	assert.True(t, exists, ".claudex/config.toml should be created")

	// Assert: config content contains defaults
	content, err := afero.ReadFile(h.FS, ".claudex/config.toml")
	require.NoError(t, err)
	assert.Contains(t, string(content), "autodoc_session_progress", "config should contain default settings")

	// Assert: sessions directory created (using app.sessionsDir which includes projectDir)
	assert.NotEmpty(t, app.sessionsDir, "sessions directory path should be set")
	exists, err = afero.DirExists(h.FS, app.sessionsDir)
	require.NoError(t, err)
	assert.True(t, exists, "sessions directory should be created")

	// Assert: logs directory created (path includes projectDir from os.Getwd())
	projectDir := app.projectDir
	logsDir := filepath.Join(projectDir, ".claudex/logs")
	exists, err = afero.DirExists(h.FS, logsDir)
	require.NoError(t, err)
	assert.True(t, exists, "logs directory should be created")

	// Assert: log file created
	assert.NotEmpty(t, app.logFilePath, "log file path should be set")
	exists, err = afero.Exists(h.FS, app.logFilePath)
	require.NoError(t, err)
	assert.True(t, exists, "log file should be created")
}

// TestInit_MigratesLegacySessions verifies legacy sessions/ folder migration
// Given: Legacy sessions/ folder with content
// When: Init() called
// Then: Sessions moved to .claudex/sessions/, old sessions/ removed
func TestInit_MigratesLegacySessions(t *testing.T) {
	// Setup - create legacy sessions/
	h := testutil.NewTestHarness()

	// Create legacy sessions directory with content
	h.WriteFile("sessions/session-1/conversation.md", "# Session 1 content")
	h.WriteFile("sessions/session-2/conversation.md", "# Session 2 content")

	// Create app
	showVersion := false
	noOverwrite := false
	updateDocs := false
	setupMCP := false
	createIndex := ""
	docPaths := []string{}

	app := &App{
		deps:            &Dependencies{FS: h.FS, Cmd: h.Commander, Clock: h, UUID: h, Env: h.Env},
		version:         "1.0.0",
		showVersion:     &showVersion,
		noOverwriteFlag: &noOverwrite,
		updateDocsFlag:  &updateDocs,
		setupMCPFlag:    &setupMCP,
		createIndexFlag: &createIndex,
		docPathsFlag:    docPaths,
	}

	h.Env.Set("HOME", "/home/user")

	// Execute Init
	err := app.Init()
	require.NoError(t, err, "Init should succeed with legacy sessions")

	// Assert: Sessions migrated to new location
	exists, err := afero.Exists(h.FS, ".claudex/sessions/session-1/conversation.md")
	require.NoError(t, err)
	assert.True(t, exists, "session-1 should be migrated to .claudex/sessions/")

	exists, err = afero.Exists(h.FS, ".claudex/sessions/session-2/conversation.md")
	require.NoError(t, err)
	assert.True(t, exists, "session-2 should be migrated to .claudex/sessions/")

	// Assert: Content preserved
	content, err := afero.ReadFile(h.FS, ".claudex/sessions/session-1/conversation.md")
	require.NoError(t, err)
	assert.Equal(t, "# Session 1 content", string(content), "session content should be preserved")

	// Assert: Old sessions/ directory removed
	exists, err = afero.DirExists(h.FS, "sessions")
	require.NoError(t, err)
	assert.False(t, exists, "legacy sessions/ directory should be removed after migration")
}

// TestInit_MigratesLegacyConfig verifies legacy .claudex.toml migration
// Given: Legacy .claudex.toml with custom values
// When: Init() called
// Then: Config moved to .claudex/config.toml, values preserved, old file removed
func TestInit_MigratesLegacyConfig(t *testing.T) {
	// Setup - create legacy config
	h := testutil.NewTestHarness()

	// Create legacy config with custom values
	legacyConfig := `# Legacy config
[features]
autodoc_session_progress = false
autodoc_session_end = false
autodoc_frequency = 10

doc = ["/custom/path"]`

	h.WriteFile(".claudex.toml", legacyConfig)

	// Create app
	showVersion := false
	noOverwrite := false
	updateDocs := false
	setupMCP := false
	createIndex := ""
	docPaths := []string{}

	app := &App{
		deps:            &Dependencies{FS: h.FS, Cmd: h.Commander, Clock: h, UUID: h, Env: h.Env},
		version:         "1.0.0",
		showVersion:     &showVersion,
		noOverwriteFlag: &noOverwrite,
		updateDocsFlag:  &updateDocs,
		setupMCPFlag:    &setupMCP,
		createIndexFlag: &createIndex,
		docPathsFlag:    docPaths,
	}

	h.Env.Set("HOME", "/home/user")

	// Execute Init
	err := app.Init()
	require.NoError(t, err, "Init should succeed with legacy config")

	// Assert: Config migrated to new location
	exists, err := afero.Exists(h.FS, ".claudex/config.toml")
	require.NoError(t, err)
	assert.True(t, exists, "config should be migrated to .claudex/config.toml")

	// Assert: Custom values preserved
	content, err := afero.ReadFile(h.FS, ".claudex/config.toml")
	require.NoError(t, err)
	assert.Contains(t, string(content), "autodoc_frequency = 10", "custom config values should be preserved")
	assert.Contains(t, string(content), "/custom/path", "custom doc paths should be preserved")

	// Assert: Old .claudex.toml removed
	exists, err = afero.Exists(h.FS, ".claudex.toml")
	require.NoError(t, err)
	assert.False(t, exists, "legacy .claudex.toml should be removed after migration")

	// Assert: Config loaded into app
	assert.NotNil(t, app.cfg, "config should be loaded into app")
}

// TestInit_LoadsConfigFromNewPath verifies config loading from .claudex/config.toml
// Given: .claudex/config.toml exists with specific values
// When: Init() called
// Then: Config values are loaded correctly into app.cfg
func TestInit_LoadsConfigFromNewPath(t *testing.T) {
	// Setup - create .claudex/ with config
	h := testutil.NewTestHarness()

	// Create config with specific values
	configContent := `# Test config
doc = ["/path/one", "/path/two"]
no_overwrite = true

[features]
autodoc_session_progress = false
autodoc_session_end = true
autodoc_frequency = 20`

	h.WriteFile(".claudex/config.toml", configContent)

	// Create app
	showVersion := false
	noOverwrite := false
	updateDocs := false
	setupMCP := false
	createIndex := ""
	docPaths := []string{}

	app := &App{
		deps:            &Dependencies{FS: h.FS, Cmd: h.Commander, Clock: h, UUID: h, Env: h.Env},
		version:         "1.0.0",
		showVersion:     &showVersion,
		noOverwriteFlag: &noOverwrite,
		updateDocsFlag:  &updateDocs,
		setupMCPFlag:    &setupMCP,
		createIndexFlag: &createIndex,
		docPathsFlag:    docPaths,
	}

	h.Env.Set("HOME", "/home/user")

	// Execute Init
	err := app.Init()
	require.NoError(t, err, "Init should succeed with existing config")

	// Assert: Config loaded into app.cfg
	assert.NotNil(t, app.cfg, "config should be loaded")

	// Assert: Config struct contains the expected values
	assert.Equal(t, []string{"/path/one", "/path/two"}, app.cfg.Doc, "config.Doc should contain doc paths")
	assert.True(t, app.cfg.NoOverwrite, "config.NoOverwrite should be true")
	assert.False(t, app.cfg.Features.AutodocSessionProgress, "config.Features.AutodocSessionProgress should be false")
	assert.True(t, app.cfg.Features.AutodocSessionEnd, "config.Features.AutodocSessionEnd should be true")
	assert.Equal(t, 20, app.cfg.Features.AutodocFrequency, "config.Features.AutodocFrequency should be 20")
}

// TestInit_MigrationIdempotent verifies migration can run multiple times safely
// Given: .claudex/ with existing content
// When: Init() called twice
// Then: No errors, no data corruption, idempotent operation
func TestInit_MigrationIdempotent(t *testing.T) {
	// Setup - create .claudex/ with content
	h := testutil.NewTestHarness()

	// Create existing session
	h.WriteFile(".claudex/sessions/existing-session/conversation.md", "# Existing content")
	h.WriteFile(".claudex/config.toml", "# Existing config\n[features]\nautodoc_frequency = 15")

	// Create app
	showVersion := false
	noOverwrite := false
	updateDocs := false
	setupMCP := false
	createIndex := ""
	docPaths := []string{}

	app := &App{
		deps:            &Dependencies{FS: h.FS, Cmd: h.Commander, Clock: h, UUID: h, Env: h.Env},
		version:         "1.0.0",
		showVersion:     &showVersion,
		noOverwriteFlag: &noOverwrite,
		updateDocsFlag:  &updateDocs,
		setupMCPFlag:    &setupMCP,
		createIndexFlag: &createIndex,
		docPathsFlag:    docPaths,
	}

	h.Env.Set("HOME", "/home/user")

	// Execute Init first time
	err := app.Init()
	require.NoError(t, err, "first Init() should succeed")

	// Read content after first init
	content1, err := afero.ReadFile(h.FS, ".claudex/sessions/existing-session/conversation.md")
	require.NoError(t, err)

	config1, err := afero.ReadFile(h.FS, ".claudex/config.toml")
	require.NoError(t, err)

	// Execute Init second time (idempotent)
	err = app.Init()
	require.NoError(t, err, "second Init() should succeed (idempotent)")

	// Assert: Content unchanged
	content2, err := afero.ReadFile(h.FS, ".claudex/sessions/existing-session/conversation.md")
	require.NoError(t, err)
	assert.Equal(t, string(content1), string(content2), "session content should be unchanged")

	config2, err := afero.ReadFile(h.FS, ".claudex/config.toml")
	require.NoError(t, err)
	assert.Equal(t, string(config1), string(config2), "config content should be unchanged")

	// Assert: No duplicate directories or corruption
	exists, err := afero.DirExists(h.FS, ".claudex")
	require.NoError(t, err)
	assert.True(t, exists, ".claudex should still exist")

	exists, err = afero.DirExists(h.FS, ".claudex/sessions")
	require.NoError(t, err)
	assert.True(t, exists, ".claudex/sessions should still exist")
}
