//go:build linux

// Package cgroup provides cgroups v2 process limiting for Linux.
// It enables true per-process-tree PID limits using the pids controller.
package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const (
	// cgroupBasePath is the default cgroups v2 mount point
	cgroupBasePath = "/sys/fs/cgroup"
	// claudexCgroupName is the parent cgroup for all claudex sessions
	claudexCgroupName = "claudex"
)

// PIDLimiter manages cgroups v2 PID limits for process trees
type PIDLimiter struct {
	mu           sync.Mutex
	basePath     string
	enabled      bool
	maxPIDs      int
	activeCgroup string // current session cgroup path
}

// NewPIDLimiter creates a new cgroups-based PID limiter.
// If cgroups v2 is not available or not writable, returns a no-op limiter.
// This is the expected behavior for non-root users outside containers.
func NewPIDLimiter(maxPIDs int) *PIDLimiter {
	limiter := &PIDLimiter{
		basePath: cgroupBasePath,
		maxPIDs:  maxPIDs,
		enabled:  false,
	}

	if maxPIDs <= 0 {
		return limiter
	}

	// Check if cgroups v2 is available
	if !isCgroupV2Available() {
		// cgroups v2 not mounted or pids controller not available
		// This is normal on non-Linux or older systems
		return limiter
	}

	// Try to create the claudex parent cgroup
	// This typically requires root or cgroup delegation (common in containers)
	claudexPath := filepath.Join(cgroupBasePath, claudexCgroupName)
	if err := os.MkdirAll(claudexPath, 0755); err != nil {
		// Can't create cgroup - expected for non-root users outside containers
		// Fall back to application-level process limiting only
		return limiter
	}

	// Enable the pids controller in the parent cgroup
	if err := enablePIDsController(claudexPath); err != nil {
		// Controller not available or not delegated to this cgroup
		// Clean up the directory we created since we can't use it
		os.Remove(claudexPath)
		return limiter
	}

	limiter.enabled = true
	return limiter
}

// IsEnabled returns true if cgroups-based limiting is active
func (l *PIDLimiter) IsEnabled() bool {
	return l.enabled
}

// CreateForProcess creates a new cgroup for a process and sets the PID limit.
// Returns the cgroup path or empty string if cgroups are not enabled.
// The caller must call Cleanup() when the process exits.
func (l *PIDLimiter) CreateForProcess(pid int) (string, error) {
	if !l.enabled || l.maxPIDs <= 0 {
		return "", nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Create a unique cgroup for this process
	cgroupName := fmt.Sprintf("cmd_%d", pid)
	cgroupPath := filepath.Join(cgroupBasePath, claudexCgroupName, cgroupName)

	// Create the cgroup directory
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create cgroup %s: %w", cgroupPath, err)
	}

	// Set the PID limit
	pidsMaxPath := filepath.Join(cgroupPath, "pids.max")
	if err := os.WriteFile(pidsMaxPath, []byte(strconv.Itoa(l.maxPIDs)), 0644); err != nil {
		// Clean up on failure
		os.Remove(cgroupPath)
		return "", fmt.Errorf("failed to set pids.max: %w", err)
	}

	// Move the process into the cgroup
	procsPath := filepath.Join(cgroupPath, "cgroup.procs")
	if err := os.WriteFile(procsPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		// Clean up on failure
		os.Remove(cgroupPath)
		return "", fmt.Errorf("failed to add process %d to cgroup: %w", pid, err)
	}

	l.activeCgroup = cgroupPath
	return cgroupPath, nil
}

// Cleanup removes a cgroup after the process has exited.
// It's safe to call even if the cgroup doesn't exist.
func (l *PIDLimiter) Cleanup(cgroupPath string) error {
	if cgroupPath == "" {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// The cgroup can only be removed when all processes have exited
	// Try to remove it - this will fail if processes are still running
	err := os.Remove(cgroupPath)
	if err != nil && !os.IsNotExist(err) {
		// If removal fails due to processes still running, that's expected
		// during cleanup - the kernel will clean up when processes exit
		if pathErr, ok := err.(*os.PathError); ok {
			if pathErr.Err == syscall.EBUSY {
				// Processes still in cgroup - will be cleaned up later
				return nil
			}
		}
		return fmt.Errorf("failed to remove cgroup %s: %w", cgroupPath, err)
	}

	if l.activeCgroup == cgroupPath {
		l.activeCgroup = ""
	}

	return nil
}

// CleanupAll removes the claudex parent cgroup and all child cgroups.
// This should be called during application shutdown.
func (l *PIDLimiter) CleanupAll() error {
	if !l.enabled {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	claudexPath := filepath.Join(cgroupBasePath, claudexCgroupName)

	// Remove all child cgroups first
	entries, err := os.ReadDir(claudexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			childPath := filepath.Join(claudexPath, entry.Name())
			os.Remove(childPath) // Ignore errors - may have active processes
		}
	}

	// Try to remove the parent cgroup
	os.Remove(claudexPath) // Ignore errors - may have active processes
	l.enabled = false

	return nil
}

// isCgroupV2Available checks if cgroups v2 is mounted and available
func isCgroupV2Available() bool {
	// Check if /sys/fs/cgroup is a cgroups v2 mount
	// In v2, there's a "cgroup.controllers" file at the root
	controllersPath := filepath.Join(cgroupBasePath, "cgroup.controllers")
	if _, err := os.Stat(controllersPath); err != nil {
		return false
	}

	// Check if pids controller is available
	data, err := os.ReadFile(controllersPath)
	if err != nil {
		return false
	}

	controllers := strings.Fields(string(data))
	for _, c := range controllers {
		if c == "pids" {
			return true
		}
	}

	return false
}

// enablePIDsController enables the pids controller in a cgroup subtree
func enablePIDsController(cgroupPath string) error {
	// To use pids controller in child cgroups, we need to enable it via subtree_control
	subtreeControlPath := filepath.Join(filepath.Dir(cgroupPath), "cgroup.subtree_control")

	// Check if pids is already enabled
	data, err := os.ReadFile(subtreeControlPath)
	if err != nil {
		return err
	}

	if strings.Contains(string(data), "pids") {
		return nil // Already enabled
	}

	// Try to enable the pids controller
	err = os.WriteFile(subtreeControlPath, []byte("+pids"), 0644)
	if err != nil {
		return fmt.Errorf("failed to enable pids controller: %w", err)
	}

	return nil
}
