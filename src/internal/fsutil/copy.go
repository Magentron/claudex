// Package fsutil provides filesystem utility functions for Claudex.
// It includes directory copying operations with support for afero.Fs
// abstraction for testability.
package fsutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// CopyDir recursively copies a directory from src to dst
func CopyDir(fs afero.Fs, src, dst string, noOverwrite bool) error {
	// Read source directory
	entries, err := afero.ReadDir(fs, src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := fs.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := CopyDir(fs, srcPath, dstPath, noOverwrite); err != nil {
				return err
			}
		} else {
			// Copy file, preserving execute permission for scripts

			// Check if noOverwrite and file exists
			if noOverwrite {
				if _, err := fs.Stat(dstPath); err == nil {
					continue // File exists, skip
				}
			}

			data, err := afero.ReadFile(fs, srcPath)
			if err != nil {
				return err
			}
			perm := os.FileMode(0644)
			if strings.HasSuffix(entry.Name(), ".sh") {
				perm = 0755
			}
			if err := afero.WriteFile(fs, dstPath, data, perm); err != nil {
				return err
			}
		}
	}

	return nil
}
