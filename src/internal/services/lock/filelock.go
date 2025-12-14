package lock

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
)

// FileLock is the production implementation of LockService
type FileLock struct {
	fs afero.Fs
}

// New creates a new LockService instance
func New(fs afero.Fs) LockService {
	return &FileLock{
		fs: fs,
	}
}

// Acquire attempts to acquire a lock at the specified path.
// Uses O_CREATE|O_EXCL flags for atomic acquisition to prevent race conditions.
// Writes the current process PID to the lock file for debugging purposes.
// Returns a Lock object if successful, or an error if the lock is already held.
func (fl *FileLock) Acquire(path string) (*Lock, error) {
	// Use O_CREATE|O_EXCL for atomic lock acquisition
	// This ensures only one process can create the file
	file, err := fl.fs.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		// Lock already exists or file operation failed
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Write current PID to lock file for debugging
	pid := os.Getpid()
	if _, err := file.Write([]byte(fmt.Sprintf("%d\n", pid))); err != nil {
		// Clean up the lock file if we can't write the PID
		file.Close()
		fl.fs.Remove(path)
		return nil, fmt.Errorf("failed to write PID to lock file: %w", err)
	}

	return &Lock{
		Path: path,
		File: file,
		fs:   fl.fs,
	}, nil
}

// IsLocked checks if a lock file exists at the given path.
// Returns true if the lock file exists, false otherwise.
func (fl *FileLock) IsLocked(path string) (bool, error) {
	_, err := fl.fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}
	return true, nil
}
