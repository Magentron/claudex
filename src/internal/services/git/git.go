// Package git provides Git repository operations for Claudex.
// It abstracts Git command execution for testability and provides
// utilities for commit history, file changes, and repository state.
package git

import (
	"strings"

	"claudex/internal/services/commander"
)

// GitService abstracts Git operations for testability
type GitService interface {
	// GetCurrentSHA returns the SHA of the current HEAD commit
	GetCurrentSHA() (string, error)

	// GetChangedFiles returns the list of changed files between base and head commits
	// Uses git diff --name-only base..head
	GetChangedFiles(base, head string) ([]string, error)

	// ValidateCommit checks if a given SHA is reachable and valid
	// Returns true if the commit exists and is accessible
	ValidateCommit(sha string) (bool, error)

	// GetMergeBase returns the merge base between HEAD and the specified branch
	// Used as fallback when base commit is unreachable (e.g., after rebase)
	GetMergeBase(branch string) (string, error)
}

// OsGitService is the production implementation of GitService
type OsGitService struct {
	cmdr commander.Commander
}

// New creates a new GitService instance
func New(cmdr commander.Commander) GitService {
	return &OsGitService{
		cmdr: cmdr,
	}
}

// GetCurrentSHA returns the SHA of the current HEAD commit
func (s *OsGitService) GetCurrentSHA() (string, error) {
	output, err := s.cmdr.Run("git", "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return trimOutput(output), nil
}

// GetChangedFiles returns the list of changed files between base and head commits
func (s *OsGitService) GetChangedFiles(base, head string) ([]string, error) {
	output, err := s.cmdr.Run("git", "diff", "--name-only", base+".."+head)
	if err != nil {
		return nil, err
	}
	return splitLines(output), nil
}

// ValidateCommit checks if a given SHA is reachable and valid
func (s *OsGitService) ValidateCommit(sha string) (bool, error) {
	output, err := s.cmdr.Run("git", "cat-file", "-t", sha)
	if err != nil {
		return false, nil
	}
	return trimOutput(output) == "commit", nil
}

// GetMergeBase returns the merge base between HEAD and the specified branch
func (s *OsGitService) GetMergeBase(branch string) (string, error) {
	output, err := s.cmdr.Run("git", "merge-base", "HEAD", branch)
	if err != nil {
		return "", err
	}
	return trimOutput(output), nil
}

// trimOutput removes leading and trailing whitespace from command output
func trimOutput(output []byte) string {
	return strings.TrimSpace(string(output))
}

// splitLines splits output by newlines and filters empty strings
func splitLines(output []byte) []string {
	lines := strings.Split(trimOutput(output), "\n")
	var result []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
