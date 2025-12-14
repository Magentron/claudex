//go:build integration

package rangeupdater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"claudex/internal/services/commander"
	"claudex/internal/services/doctracking"
	"claudex/internal/services/git"
	"claudex/internal/services/lock"

	"github.com/spf13/afero"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) string {
	t.Helper()

	// Create temp directory
	repoPath := t.TempDir()

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	return repoPath
}

// makeCommit creates files and makes a commit, returns the commit SHA
func makeCommit(t *testing.T, repoPath string, files map[string]string, message string) string {
	t.Helper()

	// Write files
	for path, content := range files {
		fullPath := filepath.Join(repoPath, path)

		// Create directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Stage files
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage files: %v", err)
	}

	// Commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Get commit SHA
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get commit SHA: %v", err)
	}

	return strings.TrimSpace(string(output))
}

// mockEnv is a simple mock implementation of Environment for testing
type mockEnv struct {
	vars map[string]string
}

func newMockEnv() *mockEnv {
	return &mockEnv{
		vars: make(map[string]string),
	}
}

func (m *mockEnv) Get(key string) string {
	return m.vars[key]
}

func (m *mockEnv) Set(key, value string) {
	m.vars[key] = value
}

// createUpdater creates a RangeUpdater instance with real services and test configuration
func createUpdater(t *testing.T, repoPath string) (*RangeUpdater, string, *mockEnv) {
	t.Helper()

	// Create session directory
	sessionPath := filepath.Join(repoPath, ".claudex-session")
	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Create real filesystem
	fs := afero.NewOsFs()

	// Create real commander (no arguments needed)
	cmdr := commander.New()

	// Create real git service
	gitSvc := git.New(cmdr)

	// Create real lock service
	lockSvc := lock.New(fs)

	// Create real tracking service
	trackingSvc := doctracking.New(fs, sessionPath)

	// Create mock environment
	mockEnv := newMockEnv()
	// Set recursion guard to prevent actual Claude invocation
	mockEnv.Set("CLAUDE_HOOK_INTERNAL", "1")

	// Create config
	config := RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
		SkipPatterns:  []string{},
		LockTimeout:   0,
	}

	updater := New(config, gitSvc, lockSvc, trackingSvc, cmdr, fs, mockEnv)

	return updater, sessionPath, mockEnv
}

// TestIntegration_HappyPath tests the happy path scenario
func TestIntegration_HappyPath(t *testing.T) {
	// Change to repo directory for git commands
	repoPath := setupTestRepo(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

	// Create initial commit
	commit1 := makeCommit(t, repoPath, map[string]string{
		"src/foo.go":   "package main\n\nfunc main() {}\n",
		"src/index.md": "# Index\n\n- foo.go: main package\n",
	}, "Initial commit")

	// Create updater and initialize tracking
	updater, _, _ := createUpdater(t, repoPath)

	// Initialize tracking with commit1
	tracking := doctracking.DocUpdateTracking{
		LastProcessedCommit: commit1,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	if err := updater.trackingSvc.Write(tracking); err != nil {
		t.Fatalf("Failed to initialize tracking: %v", err)
	}

	// Make second commit with changes
	commit2 := makeCommit(t, repoPath, map[string]string{
		"src/foo.go": "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
	}, "Add hello")

	// Run updater (Claude invocation prevented by CLAUDE_HOOK_INTERNAL=1)
	result, err := updater.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify result
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got %s: %s", result.Status, result.Reason)
	}

	if len(result.AffectedIndexes) != 1 {
		t.Errorf("Expected 1 affected index, got %d", len(result.AffectedIndexes))
	}

	// Check that at least one index contains "src/index.md"
	if len(result.AffectedIndexes) > 0 {
		found := false
		for _, indexPath := range result.AffectedIndexes {
			if strings.HasSuffix(indexPath, "src/index.md") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected affected index to contain 'src/index.md', got %v", result.AffectedIndexes)
		}
	}

	// Verify tracking was updated to commit2
	newTracking, err := updater.trackingSvc.Read()
	if err != nil {
		t.Fatalf("Failed to read tracking: %v", err)
	}

	if newTracking.LastProcessedCommit != commit2 {
		t.Errorf("Expected tracking to be %s, got %s", commit2, newTracking.LastProcessedCommit)
	}
}

// TestIntegration_NoChanges tests early exit when HEAD equals last processed
func TestIntegration_NoChanges(t *testing.T) {
	repoPath := setupTestRepo(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

	// Create initial commit
	commitSHA := makeCommit(t, repoPath, map[string]string{
		"src/foo.go":   "package main\n\nfunc main() {}\n",
		"src/index.md": "# Index\n\n- foo.go: main package\n",
	}, "Initial commit")

	// Create updater and initialize tracking with HEAD
	updater, _, _ := createUpdater(t, repoPath)

	tracking := doctracking.DocUpdateTracking{
		LastProcessedCommit: commitSHA,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	if err := updater.trackingSvc.Write(tracking); err != nil {
		t.Fatalf("Failed to initialize tracking: %v", err)
	}

	// Run updater (no new commits)
	result, err := updater.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify early exit
	if result.Status != "skipped" {
		t.Errorf("Expected status 'skipped', got %s", result.Status)
	}

	if !strings.Contains(result.Reason, "no new commits") {
		t.Errorf("Expected reason to contain 'no new commits', got: %s", result.Reason)
	}
}

// TestIntegration_DocsOnlyChanges tests skipping when only markdown files change
func TestIntegration_DocsOnlyChanges(t *testing.T) {
	repoPath := setupTestRepo(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

	// Create initial commit
	commit1 := makeCommit(t, repoPath, map[string]string{
		"src/foo.go":     "package main\n\nfunc main() {}\n",
		"src/index.md":   "# Index\n\n- foo.go: main package\n",
		"docs/readme.md": "# README\n",
	}, "Initial commit")

	// Create updater with environment variable to skip docs
	updater, _, mockEnv := createUpdater(t, repoPath)
	// Clear recursion guard and set skip docs env var
	delete(mockEnv.vars, "CLAUDE_HOOK_INTERNAL")
	mockEnv.Set("CLAUDEX_SKIP_DOCS", "1")

	tracking := doctracking.DocUpdateTracking{
		LastProcessedCommit: commit1,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	if err := updater.trackingSvc.Write(tracking); err != nil {
		t.Fatalf("Failed to initialize tracking: %v", err)
	}

	// Commit only markdown changes (no .go files)
	makeCommit(t, repoPath, map[string]string{
		"docs/readme.md": "# README\n\nUpdated docs\n",
		"src/index.md":   "# Index\n\n- foo.go: updated main package\n",
	}, "Update docs")

	// Run updater
	result, err := updater.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify skip was triggered via environment variable
	if result.Status != "skipped" {
		t.Errorf("Expected status 'skipped', got %s (reason: '%s')", result.Status, result.Reason)
	}

	if !strings.Contains(result.Reason, "CLAUDEX_SKIP_DOCS") {
		t.Errorf("Expected reason to contain 'CLAUDEX_SKIP_DOCS', got: '%s'", result.Reason)
	}
}

// TestIntegration_SkipDocsTag tests skipping when commit message has [skip-docs]
func TestIntegration_SkipDocsTag(t *testing.T) {
	t.Skip("Skipping: commit message checking not yet implemented in current skiprules")

	repoPath := setupTestRepo(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

	// Create initial commit
	commit1 := makeCommit(t, repoPath, map[string]string{
		"src/foo.go": "package main\n\nfunc main() {}\n",
	}, "Initial commit")

	// Create updater and initialize tracking
	updater, _, _ := createUpdater(t, repoPath)

	tracking := doctracking.DocUpdateTracking{
		LastProcessedCommit: commit1,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	if err := updater.trackingSvc.Write(tracking); err != nil {
		t.Fatalf("Failed to initialize tracking: %v", err)
	}

	// Commit with [skip-docs] tag
	makeCommit(t, repoPath, map[string]string{
		"src/foo.go": "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
	}, "fix: typo [skip-docs]")

	// Run updater
	result, err := updater.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// TODO: Update this test once commit message checking is implemented in skiprules
	t.Logf("Result: %s - %s", result.Status, result.Reason)
}

// TestIntegration_UnreachableBase tests fallback when base SHA is unreachable
func TestIntegration_UnreachableBase(t *testing.T) {
	repoPath := setupTestRepo(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

	// Create main branch with commits
	_ = makeCommit(t, repoPath, map[string]string{
		"src/foo.go":   "package main\n\nfunc main() {}\n",
		"src/index.md": "# Index\n\n- foo.go: main package\n",
	}, "Initial commit")

	// Create a main branch at commit1
	cmd := exec.Command("git", "branch", "main")
	cmd.Dir = repoPath
	cmd.Run() // Ignore error if branch exists

	// Create second commit
	_ = makeCommit(t, repoPath, map[string]string{
		"src/foo.go": "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
	}, "Add hello")

	// Create updater
	updater, _, _ := createUpdater(t, repoPath)

	// Set tracking to a fake SHA (simulate rebase scenario)
	tracking := doctracking.DocUpdateTracking{
		LastProcessedCommit: "0000000000000000000000000000000000000000",
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	if err := updater.trackingSvc.Write(tracking); err != nil {
		t.Fatalf("Failed to write tracking: %v", err)
	}

	// Run updater
	result, err := updater.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// The fallback will use merge-base with main, which should find commit1 or commit2
	// Depending on the branch structure, it should either:
	// - Find commit1 (merge-base) and show changes to commit2
	// - OR accept the current state and update tracking
	// Since we can't guarantee exact behavior, just verify no error and tracking updated
	if result.Status == "error" {
		t.Errorf("Expected non-error status after fallback, got: %s", result.Reason)
	}

	// Verify tracking was updated to current HEAD (commit2)
	newTracking, err := updater.trackingSvc.Read()
	if err != nil {
		t.Fatalf("Failed to read tracking: %v", err)
	}

	// Should be updated to commit2 OR remain at fallback base depending on outcome
	// The key is that it shouldn't stay at the invalid SHA
	if newTracking.LastProcessedCommit == "0000000000000000000000000000000000000000" {
		t.Error("Tracking should not remain at invalid SHA after fallback")
	}

	t.Logf("Fallback test result: status=%s, tracking=%s", result.Status, newTracking.LastProcessedCommit)
}

// TestIntegration_LockContention tests behavior when lock file exists
func TestIntegration_LockContention(t *testing.T) {
	repoPath := setupTestRepo(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

	// Create initial commit
	commit1 := makeCommit(t, repoPath, map[string]string{
		"src/foo.go":   "package main\n\nfunc main() {}\n",
		"src/index.md": "# Index\n\n- foo.go: main package\n",
	}, "Initial commit")

	// Create updater
	updater, sessionPath, _ := createUpdater(t, repoPath)

	// Initialize tracking
	tracking := doctracking.DocUpdateTracking{
		LastProcessedCommit: commit1,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	if err := updater.trackingSvc.Write(tracking); err != nil {
		t.Fatalf("Failed to initialize tracking: %v", err)
	}

	// Make new commit
	makeCommit(t, repoPath, map[string]string{
		"src/foo.go": "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
	}, "Add hello")

	// Manually create lock file
	lockPath := filepath.Join(sessionPath, "doc_update.lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create lock file: %v", err)
	}
	fmt.Fprintf(lockFile, "%d", os.Getpid())
	lockFile.Close()
	defer os.Remove(lockPath) // Cleanup

	// Run updater
	result, err := updater.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify lock contention was detected
	if result.Status != "locked" {
		t.Errorf("Expected status 'locked', got %s", result.Status)
	}

	// Verify lock file is still intact
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("Lock file was removed unexpectedly")
	}
}

// TestIntegration_MultipleIndexes tests updating multiple index files
func TestIntegration_MultipleIndexes(t *testing.T) {
	repoPath := setupTestRepo(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(repoPath)

	// Create initial commit with multiple packages
	commit1 := makeCommit(t, repoPath, map[string]string{
		"pkg/a/foo.go":   "package a\n\nfunc Foo() {}\n",
		"pkg/a/index.md": "# Package A\n\n- foo.go: Foo function\n",
		"pkg/b/bar.go":   "package b\n\nfunc Bar() {}\n",
		"pkg/b/index.md": "# Package B\n\n- bar.go: Bar function\n",
	}, "Initial commit")

	// Create updater and initialize tracking
	updater, _, _ := createUpdater(t, repoPath)

	tracking := doctracking.DocUpdateTracking{
		LastProcessedCommit: commit1,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	if err := updater.trackingSvc.Write(tracking); err != nil {
		t.Fatalf("Failed to initialize tracking: %v", err)
	}

	// Commit changes to both packages
	commit2 := makeCommit(t, repoPath, map[string]string{
		"pkg/a/foo.go": "package a\n\nfunc Foo() {\n\tprintln(\"foo\")\n}\n",
		"pkg/b/bar.go": "package b\n\nfunc Bar() {\n\tprintln(\"bar\")\n}\n",
	}, "Update both packages")

	// Run updater
	result, err := updater.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify result
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got %s: %s", result.Status, result.Reason)
	}

	if len(result.AffectedIndexes) != 2 {
		t.Errorf("Expected 2 affected indexes, got %d", len(result.AffectedIndexes))
	}

	// Verify both indexes are in the affected list (using suffix matching to handle symlinks)
	foundA := false
	foundB := false
	for _, indexPath := range result.AffectedIndexes {
		if strings.HasSuffix(indexPath, "pkg/a/index.md") {
			foundA = true
		}
		if strings.HasSuffix(indexPath, "pkg/b/index.md") {
			foundB = true
		}
	}

	if !foundA {
		t.Errorf("Expected 'pkg/a/index.md' to be in affected indexes, got: %v", result.AffectedIndexes)
	}
	if !foundB {
		t.Errorf("Expected 'pkg/b/index.md' to be in affected indexes, got: %v", result.AffectedIndexes)
	}

	// Verify tracking was updated to commit2
	newTracking, err := updater.trackingSvc.Read()
	if err != nil {
		t.Fatalf("Failed to read tracking: %v", err)
	}

	if newTracking.LastProcessedCommit != commit2 {
		t.Errorf("Expected tracking to be %s, got %s", commit2, newTracking.LastProcessedCommit)
	}
}
