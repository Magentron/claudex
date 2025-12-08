package posttooluse

import (
	"fmt"

	"claudex/internal/hooks/shared"
)

// Handler handles PostToolUse hook events.
// It logs tool completion information and always returns an "allow" decision.
type Handler struct {
	logger *shared.Logger
}

// NewHandler creates a new Handler with the provided logger.
func NewHandler(logger *shared.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// Handle processes a PostToolUse event by logging tool completion details
// and returning an "allow" permission decision.
func (h *Handler) Handle(input *shared.PostToolUseInput) (*shared.HookOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// Log tool completion with status
	logMsg := fmt.Sprintf("PostToolUse: %s completed with status %s", input.ToolName, input.Status)
	if err := h.logger.LogInfo(logMsg); err != nil {
		// Log error but don't fail the hook - logging is a side effect
		_ = h.logger.LogError(fmt.Errorf("failed to log tool completion: %w", err))
	}

	// Always return "allow" decision
	return &shared.HookOutput{
		HookSpecificOutput: shared.HookSpecificOutput{
			HookEventName:      "PostToolUse",
			PermissionDecision: "allow",
		},
	}, nil
}
