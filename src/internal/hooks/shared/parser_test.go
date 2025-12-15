package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostToolUseInput_ToolResponseTypes verifies that ToolResponse can handle different JSON types
// This test ensures the fix for MCP tools that return arrays, strings, objects, etc.
func TestPostToolUseInput_ToolResponseTypes(t *testing.T) {
	tests := []struct {
		name         string
		jsonInput    string
		wantResponse interface{}
	}{
		{
			name: "object response",
			jsonInput: `{
				"session_id": "test-123",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd": "/tmp",
				"permission_mode": "full",
				"hook_event_name": "PostToolUse",
				"tool_name": "Read",
				"tool_input": {"file_path": "/tmp/test.txt"},
				"tool_response": {"status": "success", "content": "file content"},
				"tool_use_id": "tool-1",
				"status": "success"
			}`,
			wantResponse: map[string]interface{}{
				"status":  "success",
				"content": "file content",
			},
		},
		{
			name: "array response",
			jsonInput: `{
				"session_id": "test-123",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd": "/tmp",
				"permission_mode": "full",
				"hook_event_name": "PostToolUse",
				"tool_name": "Glob",
				"tool_input": {"pattern": "*.go"},
				"tool_response": ["file1.go", "file2.go", "file3.go"],
				"tool_use_id": "tool-2",
				"status": "success"
			}`,
			wantResponse: []interface{}{"file1.go", "file2.go", "file3.go"},
		},
		{
			name: "string response",
			jsonInput: `{
				"session_id": "test-123",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd": "/tmp",
				"permission_mode": "full",
				"hook_event_name": "PostToolUse",
				"tool_name": "Bash",
				"tool_input": {"command": "echo hello"},
				"tool_response": "hello\n",
				"tool_use_id": "tool-3",
				"status": "success"
			}`,
			wantResponse: "hello\n",
		},
		{
			name: "number response",
			jsonInput: `{
				"session_id": "test-123",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd": "/tmp",
				"permission_mode": "full",
				"hook_event_name": "PostToolUse",
				"tool_name": "Calculate",
				"tool_input": {"expression": "2+2"},
				"tool_response": 4,
				"tool_use_id": "tool-4",
				"status": "success"
			}`,
			wantResponse: float64(4),
		},
		{
			name: "boolean response",
			jsonInput: `{
				"session_id": "test-123",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd": "/tmp",
				"permission_mode": "full",
				"hook_event_name": "PostToolUse",
				"tool_name": "Check",
				"tool_input": {"condition": "is_valid"},
				"tool_response": true,
				"tool_use_id": "tool-5",
				"status": "success"
			}`,
			wantResponse: true,
		},
		{
			name: "null response",
			jsonInput: `{
				"session_id": "test-123",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd": "/tmp",
				"permission_mode": "full",
				"hook_event_name": "PostToolUse",
				"tool_name": "Delete",
				"tool_input": {"path": "/tmp/test"},
				"tool_response": null,
				"tool_use_id": "tool-6",
				"status": "success"
			}`,
			wantResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input PostToolUseInput
			err := json.Unmarshal([]byte(tt.jsonInput), &input)
			require.NoError(t, err, "JSON unmarshaling should succeed for %s", tt.name)

			assert.Equal(t, tt.wantResponse, input.ToolResponse,
				"ToolResponse should match expected value for %s", tt.name)
		})
	}
}

// TestPostToolUseInput_ToolResponseMarshalRoundtrip verifies that we can marshal and unmarshal PostToolUseInput
func TestPostToolUseInput_ToolResponseMarshalRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		input    PostToolUseInput
		wantJSON string
	}{
		{
			name: "array response roundtrip",
			input: PostToolUseInput{
				HookInput: HookInput{
					SessionID:      "test-123",
					TranscriptPath: "/tmp/transcript.jsonl",
					CWD:            "/tmp",
					PermissionMode: "full",
					HookEventName:  "PostToolUse",
				},
				ToolName:     "Glob",
				ToolInput:    map[string]interface{}{"pattern": "*.go"},
				ToolResponse: []interface{}{"file1.go", "file2.go"},
				ToolUseID:    "tool-1",
				Status:       "success",
			},
			wantJSON: `{"session_id":"test-123","transcript_path":"/tmp/transcript.jsonl","cwd":"/tmp","permission_mode":"full","hook_event_name":"PostToolUse","tool_name":"Glob","tool_input":{"pattern":"*.go"},"tool_response":["file1.go","file2.go"],"tool_use_id":"tool-1","status":"success"}`,
		},
		{
			name: "string response roundtrip",
			input: PostToolUseInput{
				HookInput: HookInput{
					SessionID:      "test-123",
					TranscriptPath: "/tmp/transcript.jsonl",
					CWD:            "/tmp",
					PermissionMode: "full",
					HookEventName:  "PostToolUse",
				},
				ToolName:     "Bash",
				ToolInput:    map[string]interface{}{"command": "pwd"},
				ToolResponse: "/tmp\n",
				ToolUseID:    "tool-2",
				Status:       "success",
			},
			wantJSON: `{"session_id":"test-123","transcript_path":"/tmp/transcript.jsonl","cwd":"/tmp","permission_mode":"full","hook_event_name":"PostToolUse","tool_name":"Bash","tool_input":{"command":"pwd"},"tool_response":"/tmp\n","tool_use_id":"tool-2","status":"success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonBytes, err := json.Marshal(tt.input)
			require.NoError(t, err, "Marshaling should succeed")

			// Verify JSON matches expected
			assert.JSONEq(t, tt.wantJSON, string(jsonBytes), "Marshaled JSON should match")

			// Unmarshal back
			var roundtrip PostToolUseInput
			err = json.Unmarshal(jsonBytes, &roundtrip)
			require.NoError(t, err, "Unmarshaling should succeed")

			// Verify roundtrip preserves data
			assert.Equal(t, tt.input, roundtrip, "Roundtrip should preserve all data")
		})
	}
}
