package pretooluse

import (
	"strings"
	"testing"

	"claudex/internal/hooks/shared"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_NonTaskTool(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()
	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "test-session-123",
		},
		ToolName: "Read",
		ToolInput: map[string]interface{}{
			"file_path": "/some/path/file.txt",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	assert.Nil(t, output.HookSpecificOutput.UpdatedInput)
}

func TestHandler_NoSessionFolder(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()
	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "nonexistent-session",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Do some work",
			"description":   "Task description",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	assert.Nil(t, output.HookSpecificOutput.UpdatedInput)
}

func TestHandler_EmptyPrompt(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "",
			"description":   "Task description",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	assert.Nil(t, output.HookSpecificOutput.UpdatedInput)
}

func TestHandler_MissingPrompt(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"description":   "Task description",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	assert.Nil(t, output.HookSpecificOutput.UpdatedInput)
}

func TestHandler_SuccessfulInjection(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with files
	sessionPath := "/workspace/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create some files in the session folder
	afero.WriteFile(fs, sessionPath+"/research-topic.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/execution-plan.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/notes.txt", []byte("content"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Please analyze the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        originalPrompt,
			"description":   "Task description",
			"subagent_type": "researcher",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "PreToolUse", output.HookSpecificOutput.HookEventName)
	assert.Equal(t, "allow", output.HookSpecificOutput.PermissionDecision)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	// Verify updated input contains all original fields
	assert.Equal(t, "Task description", output.HookSpecificOutput.UpdatedInput["description"])
	assert.Equal(t, "researcher", output.HookSpecificOutput.UpdatedInput["subagent_type"])

	// Verify prompt was modified
	modifiedPrompt, ok := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	require.True(t, ok)
	assert.Contains(t, modifiedPrompt, "## SESSION CONTEXT (CRITICAL)")
	assert.Contains(t, modifiedPrompt, sessionPath)
	assert.Contains(t, modifiedPrompt, "MANDATORY RULES")
	assert.Contains(t, modifiedPrompt, "## ORIGINAL REQUEST")
	assert.Contains(t, modifiedPrompt, originalPrompt)

	// Verify session files are listed
	assert.Contains(t, modifiedPrompt, "- execution-plan.md")
	assert.Contains(t, modifiedPrompt, "- notes.txt")
	assert.Contains(t, modifiedPrompt, "- research-topic.md")
}

func TestHandler_EmptySessionFolder(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create empty session folder
	sessionPath := "/workspace/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Please analyze the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt": originalPrompt,
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	assert.Contains(t, modifiedPrompt, "### Session Folder Contents:")
	assert.Contains(t, modifiedPrompt, "(empty)")
}

func TestHandler_WithDocPaths(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)
	env.Set("CLAUDEX_DOC_PATHS", "research-{topic}.md:execution-plan-{feature}.md:analysis-{component}.md")

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Please analyze the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt": originalPrompt,
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	assert.Contains(t, modifiedPrompt, "### Recommended File Names:")
	assert.Contains(t, modifiedPrompt, "- research-{topic}.md")
	assert.Contains(t, modifiedPrompt, "- execution-plan-{feature}.md")
	assert.Contains(t, modifiedPrompt, "- analysis-{component}.md")
}

func TestHandler_PatternMatchSessionFolder(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with pattern-based name
	sessionPath := "./sessions/golang-hooks-rewrite-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create a file in the session
	afero.WriteFile(fs, sessionPath+"/research.md", []byte("content"), 0644)

	// Don't set CLAUDEX_SESSION_PATH - let it use pattern matching

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	originalPrompt := "Please analyze the codebase"
	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt": originalPrompt,
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	assert.Contains(t, modifiedPrompt, "## SESSION CONTEXT (CRITICAL)")
	assert.Contains(t, modifiedPrompt, "- research.md")
}

func TestHandler_PreservesAllToolInputFields(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder
	sessionPath := "/workspace/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt":        "Original prompt",
			"description":   "Task description",
			"subagent_type": "researcher",
			"custom_field":  "custom_value",
			"numeric_field": 42,
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	// Verify all fields are preserved
	assert.Equal(t, "Task description", output.HookSpecificOutput.UpdatedInput["description"])
	assert.Equal(t, "researcher", output.HookSpecificOutput.UpdatedInput["subagent_type"])
	assert.Equal(t, "custom_value", output.HookSpecificOutput.UpdatedInput["custom_field"])
	assert.Equal(t, 42, output.HookSpecificOutput.UpdatedInput["numeric_field"])

	// Verify only prompt is modified
	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	assert.Contains(t, modifiedPrompt, "## SESSION CONTEXT")
	assert.Contains(t, modifiedPrompt, "Original prompt")
}

func TestHandler_SessionFolderWithDirectories(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	// Create session folder with both files and directories
	sessionPath := "/workspace/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath+"/subdir", 0755)
	require.NoError(t, err)

	// Create files
	afero.WriteFile(fs, sessionPath+"/file1.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/file2.txt", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/subdir/nested.md", []byte("content"), 0644)

	env.Set("CLAUDEX_SESSION_PATH", sessionPath)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	input := &shared.PreToolUseInput{
		HookInput: shared.HookInput{
			SessionID: "abc123",
		},
		ToolName: "Task",
		ToolInput: map[string]interface{}{
			"prompt": "Test prompt",
		},
	}

	// Act
	output, err := handler.Handle(input)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, output.HookSpecificOutput.UpdatedInput)

	modifiedPrompt := output.HookSpecificOutput.UpdatedInput["prompt"].(string)
	// Should list files but not directories
	assert.Contains(t, modifiedPrompt, "- file1.md")
	assert.Contains(t, modifiedPrompt, "- file2.txt")
	assert.NotContains(t, modifiedPrompt, "- subdir")
	assert.NotContains(t, modifiedPrompt, "- nested.md") // Should not list files in subdirs
}

func TestBuildSessionContext(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	sessionPath := "/workspace/sessions/test-session-abc123"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create files in alphabetical order to test sorting
	afero.WriteFile(fs, sessionPath+"/zebra.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/apple.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/banana.md", []byte("content"), 0644)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	docPaths := []string{"research-{topic}.md", "execution-plan-{feature}.md"}

	// Act
	context, err := handler.buildSessionContext(sessionPath, docPaths)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, context, "## SESSION CONTEXT (CRITICAL)")
	assert.Contains(t, context, sessionPath)
	assert.Contains(t, context, "MANDATORY RULES")
	assert.Contains(t, context, "### Session Folder Contents:")

	// Verify files are sorted alphabetically
	appleIdx := strings.Index(context, "- apple.md")
	bananaIdx := strings.Index(context, "- banana.md")
	zebraIdx := strings.Index(context, "- zebra.md")
	assert.True(t, appleIdx < bananaIdx)
	assert.True(t, bananaIdx < zebraIdx)

	// Verify doc paths
	assert.Contains(t, context, "### Recommended File Names:")
	assert.Contains(t, context, "- research-{topic}.md")
	assert.Contains(t, context, "- execution-plan-{feature}.md")
}

func TestListSessionFiles(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	sessionPath := "/workspace/sessions/test-session"
	err := fs.MkdirAll(sessionPath, 0755)
	require.NoError(t, err)

	// Create mixed files and directories
	afero.WriteFile(fs, sessionPath+"/file1.md", []byte("content"), 0644)
	afero.WriteFile(fs, sessionPath+"/file2.txt", []byte("content"), 0644)
	fs.MkdirAll(sessionPath+"/subdir", 0755)

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	files, err := handler.listSessionFiles(sessionPath)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, len(files))
	assert.Contains(t, files, "file1.md")
	assert.Contains(t, files, "file2.txt")
	assert.NotContains(t, files, "subdir")
}

func TestListSessionFiles_NonexistentDirectory(t *testing.T) {
	// Arrange
	fs := afero.NewMemMapFs()
	env := shared.NewMockEnv()

	logger := shared.NewLogger(fs, env, "test")
	handler := NewHandler(fs, env, logger)

	// Act
	files, err := handler.listSessionFiles("/nonexistent/path")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
}
