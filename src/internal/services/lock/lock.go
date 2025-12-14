// Package lock provides file-based locking for concurrent process coordination.
// It enables cross-process synchronization using atomic file operations.
package lock

import "github.com/spf13/afero"

// Lock represents an acquired lock with its associated file handle
type Lock struct {
	// Path is the absolute path to the lock file
	Path string

	// File is the underlying file handle (for release operations)
	File afero.File

	// fs is the filesystem abstraction for cleanup
	fs afero.Fs
}

// Release removes the lock file and releases the lock
func (l *Lock) Release() error {
	if l.File != nil {
		if err := l.File.Close(); err != nil {
			return err
		}
	}
	return l.fs.Remove(l.Path)
}

// LockService abstracts file-based locking for testability
type LockService interface {
	// Acquire attempts to acquire a lock at the specified path
	// Returns Lock if successful, error if already locked or operation fails
	// Uses O_CREATE|O_EXCL for atomic acquisition
	Acquire(path string) (*Lock, error)

	// IsLocked checks if a lock file exists at the given path
	IsLocked(path string) (bool, error)
}
