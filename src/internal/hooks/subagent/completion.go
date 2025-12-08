package subagent

import (
	"fmt"

	"claudex/internal/doc"
	"claudex/internal/hooks/shared"
	"claudex/internal/notify"
	"claudex/internal/services/env"
	"claudex/internal/services/session"

	"github.com/spf13/afero"
)

// Handler implements agent completion handler with doc update and notification
type Handler struct {
	fs       afero.Fs
	env      env.Environment
	updater  doc.DocumentationUpdater
	notifier notify.Notifier
	logger   *shared.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(
	fs afero.Fs,
	env env.Environment,
	updater doc.DocumentationUpdater,
	notifier notify.Notifier,
	logger *shared.Logger,
) *Handler {
	return &Handler{
		fs:       fs,
		env:      env,
		updater:  updater,
		notifier: notifier,
		logger:   logger,
	}
}

// Handle processes subagent completion: resets counter, updates docs, and sends notification
func (h *Handler) Handle(input *shared.SubagentStopInput) (*shared.HookOutput, error) {
	_ = h.logger.LogInfo(fmt.Sprintf("Subagent stopped: %s (reason: %s)", input.AgentID, input.CompletionReason))

	// Find session folder
	sessionPath, err := session.FindSessionFolderWithCwd(h.fs, h.env, input.SessionID, input.CWD)
	if err != nil {
		// Log error but allow execution to continue
		_ = h.logger.LogError(fmt.Errorf("failed to find session folder: %w", err))
		return h.allowOutput(), nil
	}

	// Reset counter to prevent duplicate updates
	// (AutoDoc might have just run, we don't want it to run again immediately)
	if err := session.ResetCounter(h.fs, sessionPath); err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to reset counter: %w", err))
		// Continue anyway - this is not critical
	}

	_ = h.logger.LogInfo("Triggering documentation update for agent completion")

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

	// Send notification
	title := "Agent Complete"
	message := fmt.Sprintf("Agent %s finished", input.AgentID)
	sound := "Glass"

	if err := h.notifier.Send(title, message, sound); err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to send notification: %w", err))
		// Don't fail - notification is nice-to-have
	}

	return h.allowOutput(), nil
}

// allowOutput creates a standard "allow" response for SubagentStop events
func (h *Handler) allowOutput() *shared.HookOutput {
	return &shared.HookOutput{
		HookSpecificOutput: shared.HookSpecificOutput{
			HookEventName:      "SubagentStop",
			PermissionDecision: "allow",
		},
	}
}
