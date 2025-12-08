package posttooluse

import (
	"fmt"

	"claudex/internal/doc"
	"claudex/internal/hooks/shared"
	"claudex/internal/services/env"
	"claudex/internal/services/session"

	"github.com/spf13/afero"
)

// AutoDocHandler implements frequency-controlled documentation updates
type AutoDocHandler struct {
	fs        afero.Fs
	env       env.Environment
	updater   doc.DocumentationUpdater
	logger    *shared.Logger
	frequency int
}

// NewAutoDocHandler creates a new AutoDocHandler instance
func NewAutoDocHandler(fs afero.Fs, env env.Environment, updater doc.DocumentationUpdater, logger *shared.Logger, frequency int) *AutoDocHandler {
	return &AutoDocHandler{
		fs:        fs,
		env:       env,
		updater:   updater,
		logger:    logger,
		frequency: frequency,
	}
}

// Handle checks counter and triggers doc update if threshold reached
func (h *AutoDocHandler) Handle(input *shared.PostToolUseInput) (*shared.HookOutput, error) {
	// Find session folder
	sessionPath, err := session.FindSessionFolderWithCwd(h.fs, h.env, input.SessionID, input.CWD)
	if err != nil {
		// Log error but allow execution to continue
		_ = h.logger.LogError(fmt.Errorf("failed to find session folder: %w", err))
		return h.allowOutput(), nil
	}

	// Increment counter
	newCount, err := session.IncrementCounter(h.fs, sessionPath)
	if err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to increment counter: %w", err))
		return h.allowOutput(), nil
	}

	_ = h.logger.LogInfo(fmt.Sprintf("Auto-doc counter: %d/%d", newCount, h.frequency))

	// Check if we've reached the threshold
	if newCount < h.frequency {
		return h.allowOutput(), nil
	}

	// Reset counter
	if err := session.ResetCounter(h.fs, sessionPath); err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to reset counter: %w", err))
		// Continue anyway - better to update docs than to fail
	}

	_ = h.logger.LogInfo("Auto-doc threshold reached, triggering documentation update")

	// Read last processed line for incremental updates
	startLine, err := session.ReadLastProcessedLine(h.fs, sessionPath)
	if err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to read last processed line: %w", err))
		startLine = 0 // Start from beginning if we can't read the marker
	}

	// Trigger documentation update (background, non-blocking)
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

// allowOutput creates a standard "allow" response
func (h *AutoDocHandler) allowOutput() *shared.HookOutput {
	return &shared.HookOutput{
		HookSpecificOutput: shared.HookSpecificOutput{
			HookEventName:      "PostToolUse",
			PermissionDecision: "allow",
		},
	}
}
