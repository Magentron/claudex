// Package preferences provides services for managing project-level user preferences.
// It persists preferences to .claudex/preferences.json in the project directory.
package preferences

// Preferences holds project-level user preferences
type Preferences struct {
	// HookSetupDeclined indicates whether user declined git hook setup
	HookSetupDeclined bool `json:"hookSetupDeclined,omitempty"`

	// DeclinedAt is the RFC3339 timestamp when hooks were declined
	DeclinedAt string `json:"declinedAt,omitempty"`
}

// Service abstracts preferences persistence for testability
type Service interface {
	// Load reads preferences from storage
	// Returns zero-value Preferences if file doesn't exist
	Load() (Preferences, error)

	// Save persists preferences to storage atomically
	Save(prefs Preferences) error
}
