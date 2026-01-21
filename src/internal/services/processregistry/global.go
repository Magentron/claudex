package processregistry

// DefaultRegistry is the global singleton instance for centralized process tracking.
// All command execution paths should use this registry to track spawned processes.
// Initialized at package import time via init().
var DefaultRegistry ProcessRegistry

func init() {
	DefaultRegistry = NewProcessRegistry()
}
