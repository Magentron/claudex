//go:build !linux

// Package cgroup provides cgroups v2 process limiting for Linux.
// On non-Linux platforms, this is a no-op implementation.
package cgroup

// PIDLimiter manages cgroups v2 PID limits for process trees.
// On non-Linux platforms, this is a no-op.
type PIDLimiter struct {
	maxPIDs int
}

// NewPIDLimiter creates a new cgroups-based PID limiter.
// On non-Linux platforms, returns a no-op limiter.
func NewPIDLimiter(maxPIDs int) *PIDLimiter {
	return &PIDLimiter{maxPIDs: maxPIDs}
}

// IsEnabled returns true if cgroups-based limiting is active.
// Always returns false on non-Linux platforms.
func (l *PIDLimiter) IsEnabled() bool {
	return false
}

// CreateForProcess is a no-op on non-Linux platforms.
func (l *PIDLimiter) CreateForProcess(pid int) (string, error) {
	return "", nil
}

// Cleanup is a no-op on non-Linux platforms.
func (l *PIDLimiter) Cleanup(cgroupPath string) error {
	return nil
}

// CleanupAll is a no-op on non-Linux platforms.
func (l *PIDLimiter) CleanupAll() error {
	return nil
}
