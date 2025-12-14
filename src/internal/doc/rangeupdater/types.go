// Package rangeupdater provides range-based documentation update orchestration.
// It coordinates Git operations, locking, tracking, and Claude invocations
// to update index.md files based on commit range changes.
package rangeupdater

import "time"

// RangeUpdaterConfig holds configuration for the range-based doc updater
type RangeUpdaterConfig struct {
	// SessionPath is the absolute path to the Claudex session directory
	// Used for tracking file storage and lock file placement
	SessionPath string

	// DefaultBranch is the branch name to use for merge-base fallback
	// Typically "main" or "master"
	DefaultBranch string

	// SkipPatterns is a list of file path patterns to ignore
	// Files matching these patterns won't trigger doc updates
	SkipPatterns []string

	// LockTimeout is the maximum time to wait for lock acquisition
	// Zero means no waiting (immediate failure if locked)
	LockTimeout time.Duration
}
