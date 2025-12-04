package testutil

import (
	"slices"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// AssertFileExists asserts that a file exists in the filesystem
func AssertFileExists(t *testing.T, fs afero.Fs, path string) {
	t.Helper()
	exists, err := afero.Exists(fs, path)
	assert.NoError(t, err)
	assert.True(t, exists, "expected file to exist: %s", path)
}

// AssertNoFileExists asserts that a file does not exist
func AssertNoFileExists(t *testing.T, fs afero.Fs, path string) {
	t.Helper()
	exists, _ := afero.Exists(fs, path)
	assert.False(t, exists, "expected file to NOT exist: %s", path)
}

// AssertDirExists asserts that a directory exists
func AssertDirExists(t *testing.T, fs afero.Fs, path string) {
	t.Helper()
	exists, err := afero.DirExists(fs, path)
	assert.NoError(t, err)
	assert.True(t, exists, "expected directory to exist: %s", path)
}

// AssertNoDirExists asserts that a directory does not exist
func AssertNoDirExists(t *testing.T, fs afero.Fs, path string) {
	t.Helper()
	exists, _ := afero.DirExists(fs, path)
	assert.False(t, exists, "expected directory to NOT exist: %s", path)
}

// AssertFileContains asserts that a file contains the expected string
func AssertFileContains(t *testing.T, fs afero.Fs, path, expected string) {
	t.Helper()
	content, err := afero.ReadFile(fs, path)
	assert.NoError(t, err, "failed to read file: %s", path)
	assert.Contains(t, string(content), expected, "file %s should contain: %s", path, expected)
}

// AssertCommandInvoked asserts that a command was invoked with specific arguments
func AssertCommandInvoked(t *testing.T, m *MockCommander, name string, args ...string) {
	t.Helper()
	for _, inv := range m.Invocations {
		if inv.Name == name && containsAll(inv.Args, args) {
			return
		}
	}
	t.Errorf("expected command %s with args %v to be invoked", name, args)
}

// containsAll checks if all needles are present in haystack
func containsAll(haystack, needles []string) bool {
	for _, needle := range needles {
		if !slices.Contains(haystack, needle) {
			return false
		}
	}
	return true
}
