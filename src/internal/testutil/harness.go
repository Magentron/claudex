// Package testutil provides testing utilities and harnesses for Claudex.
// It includes mock implementations of dependencies and test harness setup
// for comprehensive unit testing.
package testutil

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// TestHarness provides a comprehensive test environment with mocked dependencies
type TestHarness struct {
	FS        afero.Fs
	Commander *MockCommander
	Env       *MockEnv
	UUIDs     []string
	uuidIndex int
	FixedTime time.Time
}

// NewTestHarness creates a new test harness with in-memory filesystem and mocks
func NewTestHarness() *TestHarness {
	return &TestHarness{
		FS:        afero.NewMemMapFs(),
		Commander: NewMockCommander(),
		Env:       NewMockEnv(),
		FixedTime: time.Now(),
	}
}

// New returns the next UUID from the pre-seeded list (implements UUIDGenerator)
func (h *TestHarness) New() string {
	if h.uuidIndex >= len(h.UUIDs) {
		return fmt.Sprintf("fallback-uuid-%d", h.uuidIndex)
	}
	uuid := h.UUIDs[h.uuidIndex]
	h.uuidIndex++
	return uuid
}

// Now returns the fixed timestamp (implements Clock)
func (h *TestHarness) Now() time.Time {
	return h.FixedTime
}

// CreateDir creates a directory in the test filesystem
func (h *TestHarness) CreateDir(path string) {
	h.FS.MkdirAll(path, 0755)
}

// WriteFile writes a file to the test filesystem
func (h *TestHarness) WriteFile(path, content string) {
	dir := filepath.Dir(path)
	h.FS.MkdirAll(dir, 0755)
	afero.WriteFile(h.FS, path, []byte(content), 0644)
}

// CreateSessionWithFiles creates a session directory with multiple files
func (h *TestHarness) CreateSessionWithFiles(path string, files map[string]string) {
	h.CreateDir(path)
	for name, content := range files {
		h.WriteFile(filepath.Join(path, name), content)
	}
}

// SetupConfigDir creates a configuration directory structure with files
func (h *TestHarness) SetupConfigDir(basePath string, files map[string]string) {
	for name, content := range files {
		h.WriteFile(filepath.Join(basePath, name), content)
	}
}
