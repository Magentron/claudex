package rangeupdater

import (
	"fmt"

	"claudex/internal/services/git"
)

// HandleUnreachableBase handles the case where the base commit is unreachable.
// This typically happens after a rebase or force push that rewrites history.
// It attempts to find a suitable fallback commit using merge-base with the default branch.
//
// Fallback strategy:
//  1. Try merge-base with provided defaultBranch (if not empty)
//  2. Try merge-base with "main"
//  3. Try merge-base with "master"
//  4. Return error if all attempts fail
func HandleUnreachableBase(gitSvc git.GitService, defaultBranch string) (string, error) {
	// Try provided default branch first
	if defaultBranch != "" {
		sha, err := gitSvc.GetMergeBase(defaultBranch)
		if err == nil && sha != "" {
			return sha, nil
		}
	}

	// Try "main" branch
	sha, err := gitSvc.GetMergeBase("main")
	if err == nil && sha != "" {
		return sha, nil
	}

	// Try "master" branch
	sha, err = gitSvc.GetMergeBase("master")
	if err == nil && sha != "" {
		return sha, nil
	}

	// All fallback attempts failed
	return "", fmt.Errorf("failed to find merge-base with any default branch (tried: %s, main, master)", defaultBranch)
}
