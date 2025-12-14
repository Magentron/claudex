package settings

import (
	"encoding/json"
	"testing"
)

func TestMergeSettings(t *testing.T) {
	// Define template JSON (similar to actual template structure)
	templateJSON := []byte(`{
  "permissions": {
    "allow": ["Bash(pip install:*)"],
    "deny": [],
    "ask": []
  },
  "hooks": {
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "./hooks/session-end.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-tool-use.sh"
          },
          {
            "type": "command",
            "command": ".claude/hooks/auto-doc-updater.sh"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "^(?!AskUserQuestion$).*",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/pre-tool-use.sh"
          }
        ]
      }
    ]
  }
}`)

	tests := []struct {
		name        string
		template    []byte
		existing    []byte
		validate    func(t *testing.T, result []byte, err error)
		description string
	}{
		{
			name:        "EmptyExisting",
			template:    templateJSON,
			existing:    nil,
			description: "template + nil → template content",
			validate: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				var resultSettings Settings
				if err := json.Unmarshal(result, &resultSettings); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				var templateSettings Settings
				if err := json.Unmarshal(templateJSON, &templateSettings); err != nil {
					t.Fatalf("failed to unmarshal template: %v", err)
				}

				// Should have same structure as template
				if len(resultSettings.Hooks) != len(templateSettings.Hooks) {
					t.Errorf("expected %d hook types, got %d", len(templateSettings.Hooks), len(resultSettings.Hooks))
				}

				// Verify PostToolUse has 2 hooks
				postToolUse := resultSettings.Hooks["PostToolUse"]
				if len(postToolUse) == 0 || len(postToolUse[0].Hooks) != 2 {
					t.Errorf("expected 2 PostToolUse hooks, got %d", len(postToolUse[0].Hooks))
				}
			},
		},
		{
			name:        "ExistingSubset",
			template:    templateJSON,
			description: "4 template hooks, 1 existing → adds 3 missing",
			existing: []byte(`{
  "permissions": {
    "allow": [],
    "deny": [],
    "ask": []
  },
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-tool-use.sh"
          }
        ]
      }
    ]
  }
}`),
			validate: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				var resultSettings Settings
				if err := json.Unmarshal(result, &resultSettings); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				// Should have all hook types from template
				if len(resultSettings.Hooks) != 3 {
					t.Errorf("expected 3 hook types, got %d", len(resultSettings.Hooks))
				}

				// PostToolUse should now have 2 hooks (1 existing + 1 missing)
				postToolUse := resultSettings.Hooks["PostToolUse"]
				if len(postToolUse) == 0 {
					t.Fatalf("PostToolUse hooks missing")
				}

				totalHooks := 0
				for _, entry := range postToolUse {
					totalHooks += len(entry.Hooks)
				}

				if totalHooks != 2 {
					t.Errorf("expected 2 PostToolUse hooks, got %d", totalHooks)
				}

				// Should have SessionEnd from template
				if _, ok := resultSettings.Hooks["SessionEnd"]; !ok {
					t.Error("SessionEnd hook type was not added from template")
				}

				// Should have PreToolUse from template
				if _, ok := resultSettings.Hooks["PreToolUse"]; !ok {
					t.Error("PreToolUse hook type was not added from template")
				}
			},
		},
		{
			name:        "ExistingComplete",
			template:    templateJSON,
			description: "all template hooks present → no changes",
			existing:    templateJSON, // Same as template
			validate: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				var resultSettings Settings
				if err := json.Unmarshal(result, &resultSettings); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				var templateSettings Settings
				if err := json.Unmarshal(templateJSON, &templateSettings); err != nil {
					t.Fatalf("failed to unmarshal template: %v", err)
				}

				// Should have same number of hook types
				if len(resultSettings.Hooks) != len(templateSettings.Hooks) {
					t.Errorf("expected %d hook types, got %d", len(templateSettings.Hooks), len(resultSettings.Hooks))
				}

				// PostToolUse should still have exactly 2 hooks (no duplicates)
				postToolUse := resultSettings.Hooks["PostToolUse"]
				totalHooks := 0
				for _, entry := range postToolUse {
					totalHooks += len(entry.Hooks)
				}

				if totalHooks != 2 {
					t.Errorf("expected 2 PostToolUse hooks (no duplicates), got %d", totalHooks)
				}
			},
		},
		{
			name:        "ExistingExtra",
			template:    templateJSON,
			description: "template + user hooks → preserves user hooks",
			existing: []byte(`{
  "permissions": {
    "allow": [],
    "deny": [],
    "ask": []
  },
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-tool-use.sh"
          },
          {
            "type": "command",
            "command": ".claude/hooks/custom-user-hook.sh"
          }
        ]
      }
    ],
    "CustomHookType": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "./my-custom-hook.sh"
          }
        ]
      }
    ]
  }
}`),
			validate: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				var resultSettings Settings
				if err := json.Unmarshal(result, &resultSettings); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				// Should preserve custom hook type
				if _, ok := resultSettings.Hooks["CustomHookType"]; !ok {
					t.Error("custom hook type was not preserved")
				}

				// Should preserve custom user hook in PostToolUse
				postToolUse := resultSettings.Hooks["PostToolUse"]
				hasCustomHook := false
				for _, entry := range postToolUse {
					for _, hook := range entry.Hooks {
						if hook.Command == ".claude/hooks/custom-user-hook.sh" {
							hasCustomHook = true
						}
					}
				}
				if !hasCustomHook {
					t.Error("custom user hook was not preserved")
				}

				// Should also have the missing template hook (auto-doc-updater.sh)
				hasAutoDoc := false
				for _, entry := range postToolUse {
					for _, hook := range entry.Hooks {
						if hook.Command == ".claude/hooks/auto-doc-updater.sh" {
							hasAutoDoc = true
						}
					}
				}
				if !hasAutoDoc {
					t.Error("missing template hook was not added")
				}
			},
		},
		{
			name:        "PreservesPermissions",
			template:    templateJSON,
			description: "custom permissions → permissions unchanged",
			existing: []byte(`{
  "permissions": {
    "allow": ["Bash(npm install:*)", "Write"],
    "deny": ["Read"],
    "ask": ["Edit"]
  },
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-tool-use.sh"
          }
        ]
      }
    ]
  }
}`),
			validate: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				var resultSettings Settings
				if err := json.Unmarshal(result, &resultSettings); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				// Verify permissions are unchanged
				if len(resultSettings.Permissions.Allow) != 2 {
					t.Errorf("expected 2 allow permissions, got %d", len(resultSettings.Permissions.Allow))
				}
				if len(resultSettings.Permissions.Deny) != 1 {
					t.Errorf("expected 1 deny permission, got %d", len(resultSettings.Permissions.Deny))
				}
				if len(resultSettings.Permissions.Ask) != 1 {
					t.Errorf("expected 1 ask permission, got %d", len(resultSettings.Permissions.Ask))
				}

				// Verify specific permission values
				if resultSettings.Permissions.Allow[0] != "Bash(npm install:*)" {
					t.Error("custom allow permission was not preserved")
				}
				if resultSettings.Permissions.Deny[0] != "Read" {
					t.Error("custom deny permission was not preserved")
				}
				if resultSettings.Permissions.Ask[0] != "Edit" {
					t.Error("custom ask permission was not preserved")
				}
			},
		},
		{
			name:        "Idempotent",
			template:    templateJSON,
			description: "merge twice → identical output",
			existing: []byte(`{
  "permissions": {
    "allow": [],
    "deny": [],
    "ask": []
  },
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-tool-use.sh"
          }
        ]
      }
    ]
  }
}`),
			validate: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				// Merge again with the result
				result2, err2 := MergeSettings(templateJSON, result)
				if err2 != nil {
					t.Fatalf("second merge failed: %v", err2)
				}

				// Parse both results
				var firstMerge Settings
				if err := json.Unmarshal(result, &firstMerge); err != nil {
					t.Fatalf("failed to unmarshal first merge: %v", err)
				}

				var secondMerge Settings
				if err := json.Unmarshal(result2, &secondMerge); err != nil {
					t.Fatalf("failed to unmarshal second merge: %v", err)
				}

				// Count hooks in PostToolUse for both merges
				countHooks := func(s Settings) int {
					total := 0
					for _, entry := range s.Hooks["PostToolUse"] {
						total += len(entry.Hooks)
					}
					return total
				}

				firstCount := countHooks(firstMerge)
				secondCount := countHooks(secondMerge)

				if firstCount != secondCount {
					t.Errorf("merge not idempotent: first merge has %d hooks, second has %d", firstCount, secondCount)
				}

				// Verify no duplicates after second merge
				if secondCount != 2 {
					t.Errorf("expected 2 hooks after idempotent merge, got %d", secondCount)
				}
			},
		},
		{
			name:        "DuplicateCommandPath",
			template:    templateJSON,
			description: "same command in both → no duplicate",
			existing: []byte(`{
  "permissions": {
    "allow": [],
    "deny": [],
    "ask": []
  },
  "hooks": {
    "PostToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/post-tool-use.sh"
          },
          {
            "type": "command",
            "command": ".claude/hooks/auto-doc-updater.sh"
          }
        ]
      }
    ]
  }
}`),
			validate: func(t *testing.T, result []byte, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				var resultSettings Settings
				if err := json.Unmarshal(result, &resultSettings); err != nil {
					t.Fatalf("failed to unmarshal result: %v", err)
				}

				// Count hooks and check for duplicates
				postToolUse := resultSettings.Hooks["PostToolUse"]
				commandCounts := make(map[string]int)

				for _, entry := range postToolUse {
					for _, hook := range entry.Hooks {
						commandCounts[hook.Command]++
					}
				}

				// Each command should appear exactly once
				for cmd, count := range commandCounts {
					if count != 1 {
						t.Errorf("command %s appears %d times, expected 1", cmd, count)
					}
				}

				// Should have exactly 2 unique hooks
				if len(commandCounts) != 2 {
					t.Errorf("expected 2 unique hooks, got %d", len(commandCounts))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MergeSettings(tt.template, tt.existing)
			tt.validate(t, result, err)
		})
	}
}

func TestMergeSettings_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		template    []byte
		existing    []byte
		expectError bool
		description string
	}{
		{
			name:        "InvalidTemplateJSON",
			template:    []byte(`{invalid json`),
			existing:    nil,
			expectError: true,
			description: "invalid template JSON should return error",
		},
		{
			name:        "InvalidExistingJSON",
			template:    []byte(`{"permissions":{"allow":[],"deny":[],"ask":[]},"hooks":{}}`),
			existing:    []byte(`{invalid json`),
			expectError: true,
			description: "invalid existing JSON should return error",
		},
		{
			name:        "EmptyTemplate",
			template:    []byte(``),
			existing:    []byte(`{"permissions":{"allow":[],"deny":[],"ask":[]},"hooks":{}}`),
			expectError: true,
			description: "empty template should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MergeSettings(tt.template, tt.existing)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
