package pretooluse

import (
	"fmt"
	"sort"
	"strings"

	"claudex/internal/hooks/shared"
	"claudex/internal/services/session"

	"github.com/spf13/afero"
)

// Handler processes PreToolUse hook events
// It injects session context into Task tool invocations
type Handler struct {
	fs     afero.Fs
	env    shared.Environment
	logger *shared.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(fs afero.Fs, env shared.Environment, logger *shared.Logger) *Handler {
	return &Handler{
		fs:     fs,
		env:    env,
		logger: logger,
	}
}

// Handle processes PreToolUse events
// Returns updatedInput for Task tools with session context injected
// Returns allow with no modification for non-Task tools
func (h *Handler) Handle(input *shared.PreToolUseInput) (*shared.HookOutput, error) {
	// Log the tool being invoked
	if h.logger != nil {
		_ = h.logger.Logf("Processing PreToolUse for tool: %s", input.ToolName)
	}

	// Only modify Task tool invocations
	if input.ToolName != "Task" {
		if h.logger != nil {
			_ = h.logger.Logf("Tool %s is not Task, passing through unchanged", input.ToolName)
		}
		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
			},
		}, nil
	}

	// Find session folder
	sessionPath, err := session.FindSessionFolder(h.fs, h.env, input.SessionID)
	if err != nil {
		// No session found - return allow without modification
		if h.logger != nil {
			_ = h.logger.Logf("No session folder found: %v", err)
		}
		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
			},
		}, nil
	}

	if h.logger != nil {
		_ = h.logger.Logf("Session folder found: %s", sessionPath)
	}

	// Get the original prompt
	originalPrompt, ok := input.ToolInput["prompt"].(string)
	if !ok || originalPrompt == "" {
		if h.logger != nil {
			_ = h.logger.LogInfo("No prompt found in tool_input, passing through unchanged")
		}
		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
			},
		}, nil
	}

	// Get doc paths from environment
	docPathsStr := h.env.Get("CLAUDEX_DOC_PATHS")
	var docPaths []string
	if docPathsStr != "" {
		docPaths = strings.Split(docPathsStr, ":")
	}

	// Build session context
	sessionContext, err := h.buildSessionContext(sessionPath, docPaths)
	if err != nil {
		if h.logger != nil {
			_ = h.logger.LogError(fmt.Errorf("failed to build session context: %w", err))
		}
		// On error, pass through without modification
		return &shared.HookOutput{
			HookSpecificOutput: shared.HookSpecificOutput{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
			},
		}, nil
	}

	// Build the modified prompt
	modifiedPrompt := fmt.Sprintf("%s\n\n---\n\n## ORIGINAL REQUEST\n\n%s", sessionContext, originalPrompt)

	// Create updated input with modified prompt
	updatedInput := make(map[string]interface{})
	for k, v := range input.ToolInput {
		updatedInput[k] = v
	}
	updatedInput["prompt"] = modifiedPrompt

	if h.logger != nil {
		_ = h.logger.Logf("Injected session context into Task tool prompt")
	}

	return &shared.HookOutput{
		HookSpecificOutput: shared.HookSpecificOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
			UpdatedInput:       updatedInput,
		},
	}, nil
}

// buildSessionContext creates the markdown context block
func (h *Handler) buildSessionContext(sessionPath string, docPaths []string) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("## SESSION CONTEXT (CRITICAL)\n\n")
	sb.WriteString("You are working within an active Claudex session. ")
	sb.WriteString("ALL documentation, plans, and artifacts MUST be created in the session folder.\n\n")
	sb.WriteString(fmt.Sprintf("**Session Folder (Absolute Path)**: `%s`\n\n", sessionPath))

	// Mandatory rules
	sb.WriteString("### MANDATORY RULES for Documentation:\n")
	sb.WriteString("1. ✅ ALWAYS save documentation to the session folder above\n")
	sb.WriteString("2. ✅ Use absolute paths when creating files (Write/Edit tools)\n")
	sb.WriteString("3. ✅ Before exploring the codebase, check the session folder for existing context\n")
	sb.WriteString("4. ❌ NEVER save documentation to project root or arbitrary locations\n")
	sb.WriteString("5. ❌ NEVER use relative paths for documentation files\n\n")

	// List session folder contents
	files, err := h.listSessionFiles(sessionPath)
	if err != nil {
		return "", fmt.Errorf("failed to list session files: %w", err)
	}

	sb.WriteString("### Session Folder Contents:\n")
	if len(files) == 0 {
		sb.WriteString("(empty)\n")
	} else {
		for _, file := range files {
			sb.WriteString(fmt.Sprintf("- %s\n", file))
		}
	}

	// Add doc paths if present
	if len(docPaths) > 0 {
		sb.WriteString("\n### Recommended File Names:\n")
		for _, docPath := range docPaths {
			if docPath != "" {
				sb.WriteString(fmt.Sprintf("- %s\n", docPath))
			}
		}
	}

	return sb.String(), nil
}

// listSessionFiles returns markdown list of files in session folder
func (h *Handler) listSessionFiles(sessionPath string) ([]string, error) {
	// Read directory contents
	entries, err := afero.ReadDir(h.fs, sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session directory: %w", err)
	}

	// Collect file names (exclude directories)
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	// Sort alphabetically for consistent output
	sort.Strings(files)

	return files, nil
}

// HandleFromBuilder is a convenience wrapper that returns the built output
// This is useful for command-line integration
func (h *Handler) HandleFromBuilder(input *shared.PreToolUseInput, builder *shared.Builder) error {
	output, err := h.Handle(input)
	if err != nil {
		return err
	}
	return builder.BuildCustom(*output)
}
