package doc

// DocumentationUpdater defines the interface for documentation update operations
type DocumentationUpdater interface {
	// RunBackground starts doc update in background goroutine
	// Returns immediately, update happens asynchronously
	RunBackground(config UpdaterConfig) error

	// Run executes doc update synchronously (for testing)
	Run(config UpdaterConfig) error
}
