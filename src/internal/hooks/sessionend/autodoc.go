package sessionend

import (
	"fmt"

	"claudex/internal/doc"
	"claudex/internal/hooks/shared"
	"claudex/internal/services/env"
	"claudex/internal/services/session"

	"github.com/spf13/afero"
)

// Handler implements final documentation update on session end
type Handler struct {
	fs      afero.Fs
	env     env.Environment
	updater doc.DocumentationUpdater
	logger  *shared.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(fs afero.Fs, env env.Environment, updater doc.DocumentationUpdater, logger *shared.Logger) *Handler {
	return &Handler{
		fs:      fs,
		env:     env,
		updater: updater,
		logger:  logger,
	}
}

// Handle triggers final documentation update when session ends
func (h *Handler) Handle(input *shared.SessionEndInput) (*shared.HookOutput, error) {
	_ = h.logger.LogInfo(fmt.Sprintf("Session ending: %s", input.Reason))

	// Find session folder
	sessionPath, err := session.FindSessionFolderWithCwd(h.fs, h.env, input.SessionID, input.CWD)
	if err != nil {
		// Log error but allow execution to continue
		_ = h.logger.LogError(fmt.Errorf("failed to find session folder: %w", err))
		return h.allowOutput(), nil
	}

	_ = h.logger.LogInfo("Triggering final documentation update")

	// Read last processed line for incremental updates
	startLine, err := session.ReadLastProcessedLine(h.fs, sessionPath)
	if err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to read last processed line: %w", err))
		startLine = 0 // Start from beginning if we can't read the marker
	}

	// Trigger documentation update (background, non-blocking)
	// This is the final update, so we always run it
	config := doc.UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: input.TranscriptPath,
		OutputFile:     "session-overview.md",
		PromptTemplate: "session-overview-documenter.md",
		Model:          "haiku",
		StartLine:      startLine + 1, // Start from next line (1-indexed)
	}

	if err := h.updater.RunBackground(config); err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to start background doc update: %w", err))
		// Don't fail - log and continue
	}

	return h.allowOutput(), nil
}

// allowOutput creates a standard "allow" response for SessionEnd events
func (h *Handler) allowOutput() *shared.HookOutput {
	return &shared.HookOutput{
		HookSpecificOutput: shared.HookSpecificOutput{
			HookEventName:      "SessionEnd",
			PermissionDecision: "allow",
		},
	}
}
