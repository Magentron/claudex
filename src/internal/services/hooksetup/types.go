package hooksetup

// Service defines the git hook setup interface
type Service interface {
	// IsGitRepo checks if the project directory is a git repository
	IsGitRepo() bool
	// IsInstalled checks if the claudex hook is already installed
	IsInstalled() bool
	// Install adds the claudex hook to post-commit (append-safe)
	Install() error
}
