// Package hooksetup provides Git hook installation for Claudex.
// It safely installs a post-commit hook to trigger documentation updates
// without breaking existing hooks.
package hooksetup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"claudex/internal/services/commander"
)

const (
	guardMarker = "# claudex-docs-hook"
	hookContent = `
# claudex-docs-hook
claudex --update-docs &
`
)

// FileService is the production implementation of Service
type FileService struct {
	fs         afero.Fs
	projectDir string
	cmdr       commander.Commander
}

// New creates a new Service instance
func New(fs afero.Fs, projectDir string, cmdr commander.Commander) Service {
	return &FileService{
		fs:         fs,
		projectDir: projectDir,
		cmdr:       cmdr,
	}
}

// IsGitRepo checks if .git directory exists
func (s *FileService) IsGitRepo() bool {
	info, err := s.fs.Stat(filepath.Join(s.projectDir, ".git"))
	return err == nil && info.IsDir()
}

// IsInstalled checks for guard marker in post-commit hook
func (s *FileService) IsInstalled() bool {
	hookPath := filepath.Join(s.projectDir, ".git", "hooks", "post-commit")
	data, err := afero.ReadFile(s.fs, hookPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), guardMarker)
}

// Install appends hook line to post-commit (creates if not exists)
func (s *FileService) Install() error {
	hookPath := filepath.Join(s.projectDir, ".git", "hooks", "post-commit")

	// Ensure hooks directory exists
	hooksDir := filepath.Dir(hookPath)
	if err := s.fs.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	// Check if file exists
	existing, err := afero.ReadFile(s.fs, hookPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var content string
	if len(existing) == 0 {
		// New file - add shebang
		content = "#!/bin/sh\n" + hookContent
	} else {
		// Append to existing
		content = string(existing) + "\n" + hookContent
	}

	// Write file
	if err := afero.WriteFile(s.fs, hookPath, []byte(content), 0755); err != nil {
		return err
	}

	return nil
}
