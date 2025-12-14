// Package doctracking provides tracking services for documentation update state.
// It manages persistent state for the last processed commit and strategy version.
package doctracking

// DocUpdateTracking represents the state of documentation updates
type DocUpdateTracking struct {
	// LastProcessedCommit is the SHA of the last commit that was processed
	LastProcessedCommit string `json:"last_processed_commit"`

	// UpdatedAt is the RFC3339 timestamp of when the tracking was last updated
	UpdatedAt string `json:"updated_at"`

	// StrategyVersion tracks the version of the update strategy used
	// Allows future migrations if the update logic changes
	StrategyVersion string `json:"strategy_version"`
}

// TrackingService abstracts documentation tracking persistence for testability
type TrackingService interface {
	// Read loads the current tracking state from storage
	// Returns zero-value DocUpdateTracking if file doesn't exist
	Read() (DocUpdateTracking, error)

	// Write persists the tracking state to storage atomically
	Write(tracking DocUpdateTracking) error

	// Initialize creates initial tracking state with HEAD commit
	// Used for first-time setup
	Initialize(headSHA string) error
}
