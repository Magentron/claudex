// Package notify provides notification capabilities for claudex hooks.
// It supports macOS notifications via osascript and voice synthesis via say.
package notify

import (
	"fmt"
	"runtime"
)

// Notifier provides notification capabilities for hooks.
type Notifier interface {
	// Send displays a notification with optional sound.
	// Returns nil if notifications are not available or disabled.
	Send(title, message, sound string) error

	// Speak synthesizes speech from message text.
	// Returns nil if voice synthesis is not available or disabled.
	Speak(message string) error

	// IsAvailable returns true if notifications are supported on this platform.
	IsAvailable() bool
}

// Config holds configuration for notifier initialization
type Config struct {
	// NotificationsEnabled controls whether notifications are sent (default: true)
	NotificationsEnabled bool

	// VoiceEnabled controls whether voice synthesis is used (default: false)
	VoiceEnabled bool

	// DefaultSound is the sound to use when no sound is specified (default: "default")
	DefaultSound string

	// DefaultVoice is the voice to use for speech synthesis (default: "Samantha")
	DefaultVoice string
}

// DefaultConfig returns the default notifier configuration
func DefaultConfig() Config {
	return Config{
		NotificationsEnabled: true,
		VoiceEnabled:         false,
		DefaultSound:         "default",
		DefaultVoice:         "Samantha",
	}
}

// New creates a new Notifier based on the current platform and configuration.
// On macOS, it returns a MacOSNotifier. On other platforms, it returns NoopNotifier.
func New(cfg Config, deps Dependencies) Notifier {
	if runtime.GOOS == "darwin" {
		return &macOSNotifier{
			config: cfg,
			deps:   deps,
		}
	}
	return &noopNotifier{}
}

// Dependencies contains the external dependencies needed by notifiers.
type Dependencies interface {
	// Commander provides command execution capabilities
	Commander() Commander
}

// Commander abstracts process execution for testability.
// This mirrors the interface from internal/services/commander.
type Commander interface {
	Run(name string, args ...string) ([]byte, error)
}

// NotificationTypeConfig maps notification types to titles and sounds
type NotificationTypeConfig struct {
	Title string
	Sound string
}

// DefaultNotificationTypes provides default configurations for common notification types
var DefaultNotificationTypes = map[string]NotificationTypeConfig{
	"permission_prompt": {
		Title: "Permission Required",
		Sound: "Blow",
	},
	"idle_timeout": {
		Title: "Claudex Idle",
		Sound: "Ping",
	},
	"agent_complete": {
		Title: "Agent Complete",
		Sound: "Glass",
	},
	"session_end": {
		Title: "Session Ended",
		Sound: "Tink",
	},
	"error": {
		Title: "Claudex Error",
		Sound: "Basso",
	},
}

// GetNotificationConfig returns the configuration for a notification type.
// If the type is unknown, it returns a default configuration.
func GetNotificationConfig(notificationType string) NotificationTypeConfig {
	if cfg, ok := DefaultNotificationTypes[notificationType]; ok {
		return cfg
	}
	return NotificationTypeConfig{
		Title: "Claudex",
		Sound: "default",
	}
}

// FormatSessionName extracts a human-readable session name from a session path.
// It strips the UUID suffix and converts to title case.
func FormatSessionName(sessionPath string) string {
	if sessionPath == "" {
		return "Claudex"
	}
	// This is a placeholder - actual implementation would extract the name
	// from the path and format it properly
	return "Claudex Session"
}

// ValidationError represents an error in notifier usage
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("notification validation error: %s - %s", e.Field, e.Message)
}
