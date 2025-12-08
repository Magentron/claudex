package notify

import (
	"fmt"
	"strings"
)

// macOSNotifier implements Notifier using macOS-specific tools (osascript and say).
type macOSNotifier struct {
	config Config
	deps   Dependencies
}

// Send displays a macOS notification using osascript.
// It uses AppleScript's "display notification" command to show system notifications.
func (n *macOSNotifier) Send(title, message, sound string) error {
	if !n.config.NotificationsEnabled {
		return nil
	}

	// Validate inputs
	if message == "" {
		return &ValidationError{Field: "message", Message: "message cannot be empty"}
	}

	// Use default sound if not specified
	if sound == "" {
		sound = n.config.DefaultSound
	}

	// Escape quotes in title and message for AppleScript
	title = escapeAppleScript(title)
	message = escapeAppleScript(message)
	sound = escapeAppleScript(sound)

	// Build the AppleScript command
	// Format: display notification "message" with title "title" sound name "sound"
	script := fmt.Sprintf(`display notification "%s" with title "%s" sound name "%s"`, message, title, sound)

	// Execute osascript
	output, err := n.deps.Commander().Run("osascript", "-e", script)
	if err != nil {
		// Check if osascript is not available
		if strings.Contains(err.Error(), "executable file not found") {
			// Silently ignore missing osascript - not all systems have it
			return nil
		}
		return fmt.Errorf("osascript failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// Speak synthesizes speech using the macOS say command.
func (n *macOSNotifier) Speak(message string) error {
	if !n.config.VoiceEnabled {
		return nil
	}

	// Validate input
	if message == "" {
		return &ValidationError{Field: "message", Message: "message cannot be empty"}
	}

	// Execute say command with specified voice
	// The -v flag specifies the voice to use
	output, err := n.deps.Commander().Run("say", "-v", n.config.DefaultVoice, message)
	if err != nil {
		// Check if say is not available
		if strings.Contains(err.Error(), "executable file not found") {
			// Silently ignore missing say command
			return nil
		}
		return fmt.Errorf("say command failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// IsAvailable returns true since this is the macOS-specific implementation.
func (n *macOSNotifier) IsAvailable() bool {
	return true
}

// escapeAppleScript escapes quotes and backslashes for AppleScript strings.
// AppleScript requires quotes to be escaped with backslashes.
func escapeAppleScript(s string) string {
	// Escape backslashes first, then quotes
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
