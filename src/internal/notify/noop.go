package notify

// noopNotifier is a no-operation notifier that does nothing.
// It's used on non-macOS platforms or for testing purposes.
type noopNotifier struct{}

// Send does nothing and returns nil.
func (n *noopNotifier) Send(title, message, sound string) error {
	return nil
}

// Speak does nothing and returns nil.
func (n *noopNotifier) Speak(message string) error {
	return nil
}

// IsAvailable returns false since this is a no-op implementation.
func (n *noopNotifier) IsAvailable() bool {
	return false
}

// NewNoop creates a new no-op notifier.
// This is useful for testing or when notifications are not needed.
func NewNoop() Notifier {
	return &noopNotifier{}
}
