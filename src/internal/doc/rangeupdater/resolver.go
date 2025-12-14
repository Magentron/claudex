package rangeupdater

import (
	"path/filepath"
	"sort"

	"github.com/spf13/afero"
)

// ResolveAffectedIndexes maps a list of changed files to their affected index.md files.
// It walks up the directory tree from each file to find the nearest parent index.md,
// de-duplicates the results, and returns them in sorted order for deterministic behavior.
func ResolveAffectedIndexes(fs afero.Fs, changedFiles []string) ([]string, error) {
	indexMap := make(map[string]bool)

	for _, file := range changedFiles {
		indexPath := findNearestIndexMd(fs, file)
		if indexPath != "" {
			indexMap[indexPath] = true
		}
	}

	// Convert map to sorted slice for deterministic output
	indexes := make([]string, 0, len(indexMap))
	for indexPath := range indexMap {
		indexes = append(indexes, indexPath)
	}
	sort.Strings(indexes)

	return indexes, nil
}

// findNearestIndexMd walks up the directory tree to find the nearest parent index.md.
// This is adapted from indexupdater.go:98-127 for batch processing.
func findNearestIndexMd(fs afero.Fs, filePath string) string {
	// Get absolute path and resolve any symlinks
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return ""
	}

	// Start from the file's parent directory
	dir := filepath.Dir(absPath)

	// Walk up the directory tree
	for {
		indexPath := filepath.Join(dir, "index.md")
		exists, err := afero.Exists(fs, indexPath)
		if err == nil && exists {
			return indexPath
		}

		// Check if we've reached the root
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, no index.md found
			break
		}
		dir = parent
	}

	return ""
}
