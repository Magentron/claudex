package rangeupdater

import (
	"path/filepath"
	"strings"

	"claudex/internal/services/env"
)

// ShouldSkip determines if documentation updates should be skipped based on skip rules.
// Returns (skip=true, reason) if any rule matches, (false, "") otherwise.
//
// Rules (evaluated in order):
//  1. Environment variable: CLAUDEX_SKIP_DOCS=1
//  2. Commit message contains: [skip-docs]
//  3. All changes are documentation files (*.md) - prevents infinite loops
func ShouldSkip(files []string, commitMsg string, env env.Environment) (skip bool, reason string) {
	// Rule 1: Environment variable
	if env.Get("CLAUDEX_SKIP_DOCS") == "1" {
		return true, "CLAUDEX_SKIP_DOCS environment variable is set"
	}

	// Rule 2: Commit message tag
	if strings.Contains(commitMsg, "[skip-docs]") {
		return true, "commit message contains [skip-docs] tag"
	}

	// Rule 3: All changes are markdown files (docs-only)
	if allMarkdownFiles(files) {
		return true, "all changes are documentation files (*.md) - preventing loop"
	}

	return false, ""
}

// allMarkdownFiles checks if all files in the list are markdown files
func allMarkdownFiles(files []string) bool {
	if len(files) == 0 {
		return false
	}

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		if ext != ".md" {
			return false
		}
	}

	return true
}
