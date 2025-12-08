package doc

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTranscript_AssistantMessages(t *testing.T) {
	fs := afero.NewMemMapFs()
	transcriptPath := "/test/transcript.jsonl"

	// Create test transcript with assistant messages
	content := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"I'll help you with that."}]}}
{"type":"assistant","timestamp":"2024-01-15T10:31:00Z","message":{"content":[{"type":"text","text":"Here's the solution."},{"type":"tool_use","name":"Read"}]}}
{"type":"user","timestamp":"2024-01-15T10:32:00Z","message":"User message"}
`
	afero.WriteFile(fs, transcriptPath, []byte(content), 0644)

	entries, lastLine, err := ParseTranscript(fs, transcriptPath, 1)

	require.NoError(t, err)
	assert.Equal(t, 3, lastLine)
	assert.Len(t, entries, 2)

	// First entry
	assert.Equal(t, "assistant_message", entries[0].Type)
	assert.Equal(t, "2024-01-15T10:30:00Z", entries[0].Timestamp)
	assert.Equal(t, []string{"I'll help you with that."}, entries[0].Content)

	// Second entry (should only have text content)
	assert.Equal(t, "assistant_message", entries[1].Type)
	assert.Equal(t, "2024-01-15T10:31:00Z", entries[1].Timestamp)
	assert.Equal(t, []string{"Here's the solution."}, entries[1].Content)
}

func TestParseTranscript_AgentResults(t *testing.T) {
	fs := afero.NewMemMapFs()
	transcriptPath := "/test/transcript.jsonl"

	// Create test transcript with agent results
	content := `{"type":"user","timestamp":"2024-01-15T10:30:00Z","toolUseResult":{"status":"completed","agentId":"agent-123","content":[{"type":"text","text":"Research complete."}]}}
{"type":"user","timestamp":"2024-01-15T10:31:00Z","toolUseResult":{"status":"completed","agentId":"agent-456","content":[{"type":"text","text":"Analysis done."}]}}
{"type":"user","timestamp":"2024-01-15T10:32:00Z","toolUseResult":{"status":"failed","agentId":"agent-789","content":[{"type":"text","text":"Should be filtered"}]}}
`
	afero.WriteFile(fs, transcriptPath, []byte(content), 0644)

	entries, lastLine, err := ParseTranscript(fs, transcriptPath, 1)

	require.NoError(t, err)
	assert.Equal(t, 3, lastLine)
	assert.Len(t, entries, 2) // Only completed results

	// First agent result
	assert.Equal(t, "agent_result", entries[0].Type)
	assert.Equal(t, "agent-123", entries[0].AgentID)
	assert.Equal(t, []string{"Research complete."}, entries[0].Content)

	// Second agent result
	assert.Equal(t, "agent_result", entries[1].Type)
	assert.Equal(t, "agent-456", entries[1].AgentID)
	assert.Equal(t, []string{"Analysis done."}, entries[1].Content)
}

func TestParseTranscript_StartLineOffset(t *testing.T) {
	fs := afero.NewMemMapFs()
	transcriptPath := "/test/transcript.jsonl"

	content := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"First message"}]}}
{"type":"assistant","timestamp":"2024-01-15T10:31:00Z","message":{"content":[{"type":"text","text":"Second message"}]}}
{"type":"assistant","timestamp":"2024-01-15T10:32:00Z","message":{"content":[{"type":"text","text":"Third message"}]}}
`
	afero.WriteFile(fs, transcriptPath, []byte(content), 0644)

	// Start from line 2 (should skip first message)
	entries, lastLine, err := ParseTranscript(fs, transcriptPath, 2)

	require.NoError(t, err)
	assert.Equal(t, 3, lastLine)
	assert.Len(t, entries, 2)
	assert.Equal(t, "Second message", entries[0].Content[0])
	assert.Equal(t, "Third message", entries[1].Content[0])
}

func TestParseTranscript_MalformedJSON(t *testing.T) {
	fs := afero.NewMemMapFs()
	transcriptPath := "/test/transcript.jsonl"

	// Mix valid and invalid JSON lines
	content := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"Valid message"}]}}
{invalid json line}
{"type":"assistant","timestamp":"2024-01-15T10:32:00Z","message":{"content":[{"type":"text","text":"Another valid"}]}}
`
	afero.WriteFile(fs, transcriptPath, []byte(content), 0644)

	entries, lastLine, err := ParseTranscript(fs, transcriptPath, 1)

	// Should skip malformed lines gracefully
	require.NoError(t, err)
	assert.Equal(t, 3, lastLine)
	assert.Len(t, entries, 2)
	assert.Equal(t, "Valid message", entries[0].Content[0])
	assert.Equal(t, "Another valid", entries[1].Content[0])
}

func TestParseTranscript_EmptyLines(t *testing.T) {
	fs := afero.NewMemMapFs()
	transcriptPath := "/test/transcript.jsonl"

	content := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"First"}]}}

{"type":"assistant","timestamp":"2024-01-15T10:32:00Z","message":{"content":[{"type":"text","text":"Second"}]}}
`
	afero.WriteFile(fs, transcriptPath, []byte(content), 0644)

	entries, _, err := ParseTranscript(fs, transcriptPath, 1)

	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "First", entries[0].Content[0])
	assert.Equal(t, "Second", entries[1].Content[0])
}

func TestParseTranscript_NoRelevantContent(t *testing.T) {
	fs := afero.NewMemMapFs()
	transcriptPath := "/test/transcript.jsonl"

	// Only user messages and tool uses (no assistant messages or agent results)
	content := `{"type":"user","timestamp":"2024-01-15T10:30:00Z","message":"User input"}
{"type":"tool_use","timestamp":"2024-01-15T10:31:00Z","name":"Read"}
`
	afero.WriteFile(fs, transcriptPath, []byte(content), 0644)

	entries, _, err := ParseTranscript(fs, transcriptPath, 1)

	require.NoError(t, err)
	assert.Len(t, entries, 0)
}

func TestParseTranscript_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()

	_, _, err := ParseTranscript(fs, "/nonexistent/transcript.jsonl", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open transcript")
}

func TestParseTranscript_MultipleTextContent(t *testing.T) {
	fs := afero.NewMemMapFs()
	transcriptPath := "/test/transcript.jsonl"

	// Assistant message with multiple text content blocks
	content := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":"First part."},{"type":"text","text":"Second part."}]}}`
	afero.WriteFile(fs, transcriptPath, []byte(content), 0644)

	entries, _, err := ParseTranscript(fs, transcriptPath, 1)

	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, []string{"First part.", "Second part."}, entries[0].Content)
}

func TestParseTranscript_EmptyTextContent(t *testing.T) {
	fs := afero.NewMemMapFs()
	transcriptPath := "/test/transcript.jsonl"

	// Messages with empty text should be filtered
	content := `{"type":"assistant","timestamp":"2024-01-15T10:30:00Z","message":{"content":[{"type":"text","text":""}]}}
{"type":"assistant","timestamp":"2024-01-15T10:31:00Z","message":{"content":[{"type":"text","text":"   "}]}}
{"type":"assistant","timestamp":"2024-01-15T10:32:00Z","message":{"content":[{"type":"text","text":"Valid"}]}}
`
	afero.WriteFile(fs, transcriptPath, []byte(content), 0644)

	entries, _, err := ParseTranscript(fs, transcriptPath, 1)

	require.NoError(t, err)
	assert.Len(t, entries, 1) // Only the valid message
	assert.Equal(t, "Valid", entries[0].Content[0])
}

func TestFormatTranscriptForPrompt_Empty(t *testing.T) {
	entries := []TranscriptEntry{}

	result := FormatTranscriptForPrompt(entries)

	assert.Equal(t, "No new transcript content.", result)
}

func TestFormatTranscriptForPrompt_AssistantMessages(t *testing.T) {
	entries := []TranscriptEntry{
		{
			Type:      "assistant_message",
			Timestamp: "2024-01-15T10:30:00Z",
			Content:   []string{"First message", "Second part"},
		},
	}

	result := FormatTranscriptForPrompt(entries)

	assert.Contains(t, result, "# Transcript Increment")
	assert.Contains(t, result, "## Assistant Message")
	assert.Contains(t, result, "**Timestamp**: 2024-01-15T10:30:00Z")
	assert.Contains(t, result, "First message")
	assert.Contains(t, result, "Second part")
	assert.Contains(t, result, "---")
}

func TestFormatTranscriptForPrompt_AgentResults(t *testing.T) {
	entries := []TranscriptEntry{
		{
			Type:      "agent_result",
			Timestamp: "2024-01-15T10:30:00Z",
			AgentID:   "agent-123",
			Content:   []string{"Research complete"},
		},
	}

	result := FormatTranscriptForPrompt(entries)

	assert.Contains(t, result, "## Agent Result")
	assert.Contains(t, result, "**Agent ID**: agent-123")
	assert.Contains(t, result, "Research complete")
}

func TestFormatTranscriptForPrompt_Mixed(t *testing.T) {
	entries := []TranscriptEntry{
		{
			Type:      "assistant_message",
			Timestamp: "2024-01-15T10:30:00Z",
			Content:   []string{"Assistant says hello"},
		},
		{
			Type:      "agent_result",
			Timestamp: "2024-01-15T10:31:00Z",
			AgentID:   "agent-456",
			Content:   []string{"Agent responds"},
		},
	}

	result := FormatTranscriptForPrompt(entries)

	// Check both message types are present
	assert.Contains(t, result, "## Assistant Message")
	assert.Contains(t, result, "## Agent Result")
	assert.Contains(t, result, "Assistant says hello")
	assert.Contains(t, result, "Agent responds")

	// Check proper ordering
	assistantIdx := strings.Index(result, "## Assistant Message")
	agentIdx := strings.Index(result, "## Agent Result")
	assert.Less(t, assistantIdx, agentIdx, "Assistant message should come before agent result")
}

func TestExtractTextContent(t *testing.T) {
	tests := []struct {
		name     string
		content  []rawContent
		expected []string
	}{
		{
			name: "text only",
			content: []rawContent{
				{Type: "text", Text: "Hello"},
			},
			expected: []string{"Hello"},
		},
		{
			name: "mixed types",
			content: []rawContent{
				{Type: "text", Text: "First"},
				{Type: "tool_use", Text: "should be ignored"},
				{Type: "text", Text: "Second"},
			},
			expected: []string{"First", "Second"},
		},
		{
			name: "empty text filtered",
			content: []rawContent{
				{Type: "text", Text: ""},
				{Type: "text", Text: "   "},
				{Type: "text", Text: "Valid"},
			},
			expected: []string{"Valid"},
		},
		{
			name:     "no text content",
			content:  []rawContent{{Type: "tool_use"}},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTextContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractEntry(t *testing.T) {
	tests := []struct {
		name     string
		raw      *rawTranscriptLine
		expected *TranscriptEntry
	}{
		{
			name: "valid assistant message",
			raw: &rawTranscriptLine{
				Type:      "assistant",
				Timestamp: "2024-01-15T10:30:00Z",
				Message: &rawMessage{
					Content: []rawContent{
						{Type: "text", Text: "Hello"},
					},
				},
			},
			expected: &TranscriptEntry{
				Type:      "assistant_message",
				Timestamp: "2024-01-15T10:30:00Z",
				Content:   []string{"Hello"},
			},
		},
		{
			name: "assistant message without content",
			raw: &rawTranscriptLine{
				Type:      "assistant",
				Timestamp: "2024-01-15T10:30:00Z",
				Message: &rawMessage{
					Content: []rawContent{},
				},
			},
			expected: nil,
		},
		{
			name: "valid agent result",
			raw: &rawTranscriptLine{
				Type:      "user",
				Timestamp: "2024-01-15T10:30:00Z",
				ToolUseResult: &rawToolUseResult{
					Status:  "completed",
					AgentID: "agent-123",
					Content: []rawContent{
						{Type: "text", Text: "Done"},
					},
				},
			},
			expected: &TranscriptEntry{
				Type:      "agent_result",
				Timestamp: "2024-01-15T10:30:00Z",
				AgentID:   "agent-123",
				Content:   []string{"Done"},
			},
		},
		{
			name: "agent result without agent ID",
			raw: &rawTranscriptLine{
				Type:      "user",
				Timestamp: "2024-01-15T10:30:00Z",
				ToolUseResult: &rawToolUseResult{
					Status:  "completed",
					AgentID: "",
					Content: []rawContent{
						{Type: "text", Text: "Should be filtered"},
					},
				},
			},
			expected: nil,
		},
		{
			name: "agent result not completed",
			raw: &rawTranscriptLine{
				Type:      "user",
				Timestamp: "2024-01-15T10:30:00Z",
				ToolUseResult: &rawToolUseResult{
					Status:  "failed",
					AgentID: "agent-123",
					Content: []rawContent{
						{Type: "text", Text: "Should be filtered"},
					},
				},
			},
			expected: nil,
		},
		{
			name: "unrelated type",
			raw: &rawTranscriptLine{
				Type:      "other",
				Timestamp: "2024-01-15T10:30:00Z",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractEntry(tt.raw)
			assert.Equal(t, tt.expected, result)
		})
	}
}
