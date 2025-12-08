package notification

import (
	"fmt"

	"claudex/internal/hooks/shared"
	"claudex/internal/notify"
)

// Handler handles notification hook events.
// It sends notifications via the configured notifier and optionally speaks messages.
type Handler struct {
	notifier notify.Notifier
	logger   *shared.Logger
	env      shared.Environment
}

// NewHandler creates a new Handler with the provided dependencies.
func NewHandler(notifier notify.Notifier, logger *shared.Logger, env shared.Environment) *Handler {
	return &Handler{
		notifier: notifier,
		logger:   logger,
		env:      env,
	}
}

// Handle processes a notification event by sending a notification and optionally speaking the message.
// It maps notification types to appropriate titles and sounds.
// Returns nil error on success (no JSON output is needed for notifications).
func (h *Handler) Handle(input *shared.NotificationInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	if input.Message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	// Log notification processing
	logMsg := fmt.Sprintf("Processing notification: type=%s, message=%s", input.NotificationType, input.Message)
	if err := h.logger.LogInfo(logMsg); err != nil {
		// Log error but continue - logging is a side effect
		_ = h.logger.LogError(fmt.Errorf("failed to log notification: %w", err))
	}

	// Get configuration for this notification type
	config := notify.GetNotificationConfig(input.NotificationType)

	// Send notification
	if err := h.notifier.Send(config.Title, input.Message, config.Sound); err != nil {
		// Log error and return - notification failure is a real error
		logErr := h.logger.LogError(fmt.Errorf("failed to send notification: %w", err))
		_ = logErr // Ignore logging errors
		return fmt.Errorf("failed to send notification: %w", err)
	}

	// Check if voice is enabled
	voiceEnabled := h.env.Get("CLAUDEX_VOICE_ENABLED")
	if voiceEnabled == "true" || voiceEnabled == "1" {
		if err := h.notifier.Speak(input.Message); err != nil {
			// Voice synthesis failure is logged but doesn't fail the hook
			logErr := h.logger.LogError(fmt.Errorf("failed to speak message: %w", err))
			_ = logErr // Ignore logging errors
		}
	}

	return nil
}
