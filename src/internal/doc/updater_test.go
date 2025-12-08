package doc

import (
	"testing"

	"claudex/internal/testutil"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdater(t *testing.T) {
	h := testutil.NewTestHarness()

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	assert.NotNil(t, updater)
	assert.Equal(t, h.FS, updater.fs)
	assert.Equal(t, h.Commander, updater.cmd)
	assert.Equal(t, h.Env, updater.env)
}

func TestRun_Success(t *testing.T) {
	t.Skip("Skipping test that requires real claude command - tested in integration tests")

	h := testutil.NewTestHarness()

	// Setup test files
	sessionPath := "/test/session"
	transcriptPath := "/test/transcript.jsonl"
	templatePath := "/test/template.md"

	h.CreateDir(sessionPath)

	// Create transcript with test data
	transcript := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"Hello world"}]}}
`
	h.WriteFile(transcriptPath, transcript)

	// Create template
	template := "Content: $RELEVANT_CONTENT\nContext: $DOC_CONTEXT"
	h.WriteFile(templatePath, template)

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: transcriptPath,
		PromptTemplate: templatePath,
		SessionContext: "Test context",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.Run(config)

	require.NoError(t, err)

	// Verify last processed line was updated
	lastLineData, err := afero.ReadFile(h.FS, sessionPath+"/.last-processed-line-overview")
	require.NoError(t, err)
	assert.Equal(t, "1", string(lastLineData))
}

func TestRun_RecursionGuard(t *testing.T) {
	h := testutil.NewTestHarness()

	// Set recursion guard
	h.Env.Set("CLAUDE_HOOK_INTERNAL", "1")

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    "/test/session",
		TranscriptPath: "/test/transcript.jsonl",
		PromptTemplate: "/test/template.md",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.Run(config)

	// Should fail due to recursion guard
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recursion guard")

	// Claude should not be called
	assert.Len(t, h.Commander.Invocations, 0)
}

func TestRun_NoNewContent(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/test/session"
	transcriptPath := "/test/transcript.jsonl"
	templatePath := "/test/template.md"

	h.CreateDir(sessionPath)

	// Create transcript with only filtered content
	transcript := `{"type":"user","timestamp":"2024-01-15T10:30:00Z","message":"User message"}
`
	h.WriteFile(transcriptPath, transcript)
	h.WriteFile(templatePath, "Template")

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: transcriptPath,
		PromptTemplate: templatePath,
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.Run(config)

	// Should succeed but not call claude
	require.NoError(t, err)
	assert.Len(t, h.Commander.Invocations, 0)
}

func TestRun_TranscriptNotFound(t *testing.T) {
	h := testutil.NewTestHarness()

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    "/test/session",
		TranscriptPath: "/nonexistent.jsonl",
		PromptTemplate: "/test/template.md",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.Run(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse transcript")
}

func TestRun_TemplateNotFound(t *testing.T) {
	h := testutil.NewTestHarness()

	transcriptPath := "/test/transcript.jsonl"
	transcript := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"Hello"}]}}
`
	h.WriteFile(transcriptPath, transcript)

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    "/test/session",
		TranscriptPath: transcriptPath,
		PromptTemplate: "/nonexistent.md",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.Run(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load prompt template")
}

func TestRun_PromptBuilding(t *testing.T) {
	t.Skip("Skipping test that requires real claude command - prompt building tested separately")

	h := testutil.NewTestHarness()

	sessionPath := "/test/session"
	transcriptPath := "/test/transcript.jsonl"
	templatePath := "/test/template.md"

	h.CreateDir(sessionPath)

	// Create transcript
	transcript := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"Test content"}]}}
`
	h.WriteFile(transcriptPath, transcript)

	// Template with placeholders
	template := "Transcript:\n$RELEVANT_CONTENT\n\nContext:\n$DOC_CONTEXT"
	h.WriteFile(templatePath, template)

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: transcriptPath,
		PromptTemplate: templatePath,
		SessionContext: "Session context here",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.Run(config)
	require.NoError(t, err)
}

func TestRunBackground_Success(t *testing.T) {
	h := testutil.NewTestHarness()

	sessionPath := "/test/session"
	transcriptPath := "/test/transcript.jsonl"
	templatePath := "/test/template.md"

	h.CreateDir(sessionPath)
	h.WriteFile(transcriptPath, `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"Hello"}]}}`)
	h.WriteFile(templatePath, "Template: $RELEVANT_CONTENT")

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: transcriptPath,
		PromptTemplate: templatePath,
		Model:          "haiku",
		StartLine:      1,
	}

	// RunBackground should return immediately
	err := updater.RunBackground(config)

	require.NoError(t, err)
	// Note: We can't reliably test background execution completion in unit tests
	// The goroutine may or may not have completed by the time we check
}

func TestValidateConfig_AllValid(t *testing.T) {
	h := testutil.NewTestHarness()
	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    "/test/session",
		TranscriptPath: "/test/transcript.jsonl",
		PromptTemplate: "/test/template.md",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.validateConfig(config)

	assert.NoError(t, err)
}

func TestValidateConfig_MissingSessionPath(t *testing.T) {
	h := testutil.NewTestHarness()
	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		TranscriptPath: "/test/transcript.jsonl",
		PromptTemplate: "/test/template.md",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.validateConfig(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SessionPath is required")
}

func TestValidateConfig_MissingTranscriptPath(t *testing.T) {
	h := testutil.NewTestHarness()
	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    "/test/session",
		PromptTemplate: "/test/template.md",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.validateConfig(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TranscriptPath is required")
}

func TestValidateConfig_MissingPromptTemplate(t *testing.T) {
	h := testutil.NewTestHarness()
	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    "/test/session",
		TranscriptPath: "/test/transcript.jsonl",
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.validateConfig(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PromptTemplate is required")
}

func TestValidateConfig_MissingModel(t *testing.T) {
	h := testutil.NewTestHarness()
	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    "/test/session",
		TranscriptPath: "/test/transcript.jsonl",
		PromptTemplate: "/test/template.md",
		StartLine:      1,
	}

	err := updater.validateConfig(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Model is required")
}

func TestValidateConfig_InvalidStartLine(t *testing.T) {
	h := testutil.NewTestHarness()
	updater := NewUpdater(h.FS, h.Commander, h.Env)

	tests := []struct {
		name      string
		startLine int
	}{
		{"zero", 0},
		{"negative", -1},
		{"large negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := UpdaterConfig{
				SessionPath:    "/test/session",
				TranscriptPath: "/test/transcript.jsonl",
				PromptTemplate: "/test/template.md",
				Model:          "haiku",
				StartLine:      tt.startLine,
			}

			err := updater.validateConfig(config)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "StartLine must be >= 1")
		})
	}
}

func TestRun_IncrementalProcessing(t *testing.T) {
	t.Skip("Skipping test that requires real claude command")

	h := testutil.NewTestHarness()

	sessionPath := "/test/session"
	transcriptPath := "/test/transcript.jsonl"
	templatePath := "/test/template.md"

	h.CreateDir(sessionPath)

	// Create transcript with multiple lines
	transcript := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"First"}]}}
{"type":"assistant","timestamp":"2024-01-15T10:31:00Z","message":{"content":[{"type":"text","text":"Second"}]}}
{"type":"assistant","timestamp":"2024-01-15T10:32:00Z","message":{"content":[{"type":"text","text":"Third"}]}}
`
	h.WriteFile(transcriptPath, transcript)
	h.WriteFile(templatePath, "Template: $RELEVANT_CONTENT")

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	// First run: process from line 1
	config := UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: transcriptPath,
		PromptTemplate: templatePath,
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.Run(config)
	require.NoError(t, err)

	// Check last processed line
	lastLineData, err := afero.ReadFile(h.FS, sessionPath+"/.last-processed-line-overview")
	require.NoError(t, err)
	assert.Equal(t, "3", string(lastLineData))

	// Second run: process from line 4 (should have no new content)
	config.StartLine = 4

	err = updater.Run(config)
	require.NoError(t, err)
}

func TestRun_LastProcessedLineUpdate(t *testing.T) {
	t.Skip("Skipping test that requires real claude command")

	h := testutil.NewTestHarness()

	sessionPath := "/test/session"
	transcriptPath := "/test/transcript.jsonl"
	templatePath := "/test/template.md"

	h.CreateDir(sessionPath)

	transcript := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"Content"}]}}
`
	h.WriteFile(transcriptPath, transcript)
	h.WriteFile(templatePath, "Template")

	updater := NewUpdater(h.FS, h.Commander, h.Env)

	config := UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: transcriptPath,
		PromptTemplate: templatePath,
		Model:          "haiku",
		StartLine:      1,
	}

	err := updater.Run(config)
	require.NoError(t, err)

	// Verify file exists and contains correct line number
	exists, err := afero.Exists(h.FS, sessionPath+"/.last-processed-line-overview")
	require.NoError(t, err)
	assert.True(t, exists)

	content, err := afero.ReadFile(h.FS, sessionPath+"/.last-processed-line-overview")
	require.NoError(t, err)
	assert.Equal(t, "1", string(content))
}
