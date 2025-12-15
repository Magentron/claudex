package doc

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPromptTemplate_Success(t *testing.T) {
	fs := afero.NewMemMapFs()
	templatePath := "/templates/prompt.md"
	content := "# Prompt Template\n\nThis is a test template."

	afero.WriteFile(fs, templatePath, []byte(content), 0644)

	result, err := LoadPromptTemplate(fs, templatePath)

	require.NoError(t, err)
	assert.Equal(t, content, result)
}

func TestLoadPromptTemplate_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()

	_, err := LoadPromptTemplate(fs, "/nonexistent/template.md")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read prompt template")
}

func TestLoadPromptTemplate_EmptyFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	templatePath := "/templates/empty.md"

	afero.WriteFile(fs, templatePath, []byte(""), 0644)

	_, err := LoadPromptTemplate(fs, templatePath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt template is empty")
}

func TestLoadPromptTemplate_WhitespaceOnly(t *testing.T) {
	fs := afero.NewMemMapFs()
	templatePath := "/templates/whitespace.md"

	afero.WriteFile(fs, templatePath, []byte("   \n\n   \t\t  "), 0644)

	_, err := LoadPromptTemplate(fs, templatePath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt template is empty")
}

func TestBuildDocumentationPrompt_BasicReplacement(t *testing.T) {
	template := "Transcript:\n$RELEVANT_CONTENT\n\nContext:\n$DOC_CONTEXT"
	transcriptContent := "Assistant said hello."
	sessionContext := "Session folder: /test/session"

	result := BuildDocumentationPrompt(template, transcriptContent, sessionContext, "/test/session")

	expected := "Transcript:\nAssistant said hello.\n\nContext:\nSession folder: /test/session"
	assert.Equal(t, expected, result)
}

func TestBuildDocumentationPrompt_MultiplePlaceholders(t *testing.T) {
	template := "$RELEVANT_CONTENT\n---\n$RELEVANT_CONTENT\n$DOC_CONTEXT"
	transcriptContent := "Content"
	sessionContext := "Context"

	result := BuildDocumentationPrompt(template, transcriptContent, sessionContext, "/test/session")

	// Both occurrences of $RELEVANT_CONTENT should be replaced
	expected := "Content\n---\nContent\nContext"
	assert.Equal(t, expected, result)
}

func TestBuildDocumentationPrompt_NoPlaceholders(t *testing.T) {
	template := "Static template with no variables"
	transcriptContent := "Content"
	sessionContext := "Context"

	result := BuildDocumentationPrompt(template, transcriptContent, sessionContext, "/test/session")

	assert.Equal(t, template, result)
}

func TestBuildDocumentationPrompt_EmptyContent(t *testing.T) {
	template := "Transcript: $RELEVANT_CONTENT | Context: $DOC_CONTEXT"
	transcriptContent := ""
	sessionContext := ""

	result := BuildDocumentationPrompt(template, transcriptContent, sessionContext, "/test/session")

	expected := "Transcript:  | Context: "
	assert.Equal(t, expected, result)
}

func TestBuildDocumentationPrompt_RealWorldExample(t *testing.T) {
	template := `# Session Documentation Update

## Transcript Content
$RELEVANT_CONTENT

## Session Context
$DOC_CONTEXT

## Task
Update session-overview.md with new information.`

	transcriptContent := "## Assistant Message\nI implemented feature X.\n\n---"
	sessionContext := "Files:\n- research.md\n- plan.md"

	result := BuildDocumentationPrompt(template, transcriptContent, sessionContext, "/test/session")

	assert.Contains(t, result, "# Session Documentation Update")
	assert.Contains(t, result, "## Assistant Message")
	assert.Contains(t, result, "I implemented feature X.")
	assert.Contains(t, result, "Files:\n- research.md\n- plan.md")
	assert.Contains(t, result, "Update session-overview.md")
}

func TestBuildDocumentationPrompt_SpecialCharacters(t *testing.T) {
	template := "Content: $RELEVANT_CONTENT"
	transcriptContent := "Special chars: $100, $test, $$"

	result := BuildDocumentationPrompt(template, transcriptContent, "", "/test/session")

	// Should handle $ characters in content without treating them as placeholders
	assert.Equal(t, "Content: Special chars: $100, $test, $$", result)
}

func TestBuildDocumentationPrompt_Multiline(t *testing.T) {
	template := "START\n$RELEVANT_CONTENT\nEND"
	transcriptContent := "Line 1\nLine 2\nLine 3"

	result := BuildDocumentationPrompt(template, transcriptContent, "", "/test/session")

	expected := "START\nLine 1\nLine 2\nLine 3\nEND"
	assert.Equal(t, expected, result)
}

// TestBuildDocumentationPrompt_ReplacesSessionFolder tests that $SESSION_FOLDER placeholder is replaced
func TestBuildDocumentationPrompt_ReplacesSessionFolder(t *testing.T) {
	// Template contains $SESSION_FOLDER placeholder (as used in session-overview-documenter.md)
	template := "Session folder: $SESSION_FOLDER\nTranscript: $RELEVANT_CONTENT\nContext: $DOC_CONTEXT"
	transcriptContent := "Assistant did work."
	sessionContext := "Existing docs"
	sessionFolder := "/Users/test/.claudex/sessions/test-123"

	result := BuildDocumentationPrompt(template, transcriptContent, sessionContext, sessionFolder)

	// Verify $SESSION_FOLDER is replaced
	assert.NotContains(t, result, "$SESSION_FOLDER",
		"Expected $SESSION_FOLDER placeholder to be replaced, but it remains in output")
	assert.Contains(t, result, sessionFolder, "Expected result to contain the session folder path")

	// Verify other placeholders ARE replaced
	assert.Contains(t, result, "Assistant did work.")
	assert.Contains(t, result, "Existing docs")
}
