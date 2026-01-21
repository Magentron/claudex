// Package processregistry provides thread-safe tracking of spawned child process PIDs.
// It enables centralized process lifecycle management for runaway process protection.
package processregistry

import "sync"

// ProcessRegistry abstracts process ID tracking for testability and centralized lifecycle management.
// All methods are thread-safe for concurrent access from multiple goroutines.
type ProcessRegistry interface {
	// Register adds a process ID to the tracking registry.
	// This should be called immediately after successfully spawning a child process.
	// Thread-safe for concurrent registration.
	Register(pid int)

	// Unregister removes a process ID from the tracking registry.
	// This should be called after a process exits or is terminated.
	// Thread-safe for concurrent unregistration.
	// Unregistering a non-existent PID is a no-op.
	Unregister(pid int)

	// Count returns the current number of tracked process IDs.
	// Thread-safe for concurrent reads.
	Count() int

	// GetAll returns a snapshot of all currently tracked process IDs.
	// The returned slice is a copy and safe to modify.
	// Thread-safe for concurrent reads.
	GetAll() []int
}

// processRegistry is the production implementation of ProcessRegistry.
// Uses sync.RWMutex for thread-safe concurrent access with optimized read performance.
type processRegistry struct {
	pids map[int]bool
	mu   sync.RWMutex
}

// NewProcessRegistry creates a new ProcessRegistry instance with an empty tracking map.
func NewProcessRegistry() ProcessRegistry {
	return &processRegistry{
		pids: make(map[int]bool),
	}
}

// Register adds a process ID to the tracking registry with write lock protection.
func (r *processRegistry) Register(pid int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pids[pid] = true
}

// Unregister removes a process ID from the tracking registry with write lock protection.
// Removing a non-existent PID is a no-op and does not cause an error.
func (r *processRegistry) Unregister(pid int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.pids, pid)
}

// Count returns the current number of tracked PIDs with read lock protection.
func (r *processRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.pids)
}

// GetAll returns a snapshot copy of all tracked PIDs with read lock protection.
// The returned slice is safe to modify without affecting the registry.
func (r *processRegistry) GetAll() []int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]int, 0, len(r.pids))
	for pid := range r.pids {
		result = append(result, pid)
	}
	return result
}
