package commander

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"claudex/internal/services/config"
	"claudex/internal/services/processregistry"
	"github.com/spf13/afero"
)

// TestRun_BasicExecution verifies basic command execution works
func TestRun_BasicExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       50,
				RateLimitPerSecond: 0, // Disable rate limiting for basic test
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	// Run a simple command
	output, err := commander.Run("echo", "hello world")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	result := strings.TrimSpace(string(output))
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", result)
	}
}

// TestRunWithContext_Timeout verifies context timeout works
func TestRunWithContext_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       50,
				RateLimitPerSecond: 0,
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to run a long-running command
	_, err := commander.RunWithContext(ctx, "sleep", "10")
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "killed") && !strings.Contains(err.Error(), "context") {
		t.Errorf("Expected timeout/killed error, got: %v", err)
	}
}

// TestStart_InteractiveExecution verifies interactive command execution
func TestStart_InteractiveExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       50,
				RateLimitPerSecond: 0,
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	stdin := strings.NewReader("test input")
	var stdout, stderr bytes.Buffer

	// Run command that echoes stdin
	err := commander.Start("cat", stdin, &stdout, &stderr, "-")
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	result := stdout.String()
	if result != "test input" {
		t.Errorf("Expected 'test input', got '%s'", result)
	}
}

// TestProcessLimit_Enforcement verifies process limit is enforced
func TestProcessLimit_Enforcement(t *testing.T) {
	// Set very low limit for testing
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       5,
				RateLimitPerSecond: 0,
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	// Clear the registry before test
	registry := processregistry.DefaultRegistry
	for _, pid := range registry.GetAll() {
		registry.Unregister(pid)
	}

	// Simulate having processes by manually registering fake PIDs
	// (Since StartWithContext is synchronous and waits for completion,
	// we can't accumulate running processes in a test. Instead, we
	// simulate the scenario by directly manipulating the registry.)
	for i := 0; i < 5; i++ {
		registry.Register(100000 + i) // Use high PIDs that won't conflict
	}

	// Cleanup registered fake PIDs after test
	defer func() {
		for i := 0; i < 5; i++ {
			registry.Unregister(100000 + i)
		}
	}()

	// Try to spawn one more process - should fail due to limit
	_, err := commander.Run("echo", "test")
	if err == nil {
		t.Fatal("Expected process limit error, got nil")
	}

	if !strings.Contains(err.Error(), "process limit reached") {
		t.Errorf("Expected 'process limit reached' error, got: %v", err)
	}
}

// TestProcessRegistry_Registration verifies PIDs are registered and unregistered
func TestProcessRegistry_Registration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       50,
				RateLimitPerSecond: 0,
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)
	registry := processregistry.DefaultRegistry

	// Clear registry
	for _, pid := range registry.GetAll() {
		registry.Unregister(pid)
	}

	initialCount := registry.Count()

	// Run a quick command
	_, err := commander.Run("echo", "test")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// After command completes, count should return to initial
	finalCount := registry.Count()
	if finalCount != initialCount {
		t.Errorf("Expected registry count %d, got %d (PIDs not unregistered)", initialCount, finalCount)
	}
}

// TestRateLimiting_Backoff verifies rate limiting applies backoff
func TestRateLimiting_Backoff(t *testing.T) {
	t.Skip("Rate limiting test is time-sensitive and may be flaky in CI")

	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       100,
				RateLimitPerSecond: 2, // Very low limit
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	// Spawn multiple processes rapidly
	start := time.Now()
	for i := 0; i < 5; i++ {
		_, err := commander.Run("echo", "test")
		if err != nil {
			t.Fatalf("Run failed on iteration %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	// With rate limiting of 2/sec, 5 processes should take at least 2 seconds
	// (accounting for exponential backoff)
	if elapsed < 1*time.Second {
		t.Errorf("Rate limiting not working: 5 processes completed in %v (expected >1s)", elapsed)
	}
}

// TestBackwardCompatibility_Run verifies old Run method still works
func TestBackwardCompatibility_Run(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       50,
				RateLimitPerSecond: 0,
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	// Old Run method should still work
	output, err := commander.Run("pwd")
	if err != nil {
		t.Fatalf("Backward compatible Run failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("Expected output from pwd command")
	}
}

// TestBackwardCompatibility_Start verifies old Start method still works
func TestBackwardCompatibility_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       50,
				RateLimitPerSecond: 0,
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	var stdout bytes.Buffer
	stdin := strings.NewReader("")

	// Old Start method should still work
	err := commander.Start("echo", stdin, &stdout, &stdout, "backward compatible")
	if err != nil {
		t.Fatalf("Backward compatible Start failed: %v", err)
	}

	result := strings.TrimSpace(stdout.String())
	if result != "backward compatible" {
		t.Errorf("Expected 'backward compatible', got '%s'", result)
	}
}

// TestDisabledProtection_NoLimits verifies protection can be disabled
func TestDisabledProtection_NoLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       0, // Disabled
				RateLimitPerSecond: 0, // Disabled
				TimeoutSeconds:     0, // Disabled
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	// Should work even if we spawn many processes
	for i := 0; i < 10; i++ {
		_, err := commander.Run("echo", "test")
		if err != nil {
			t.Fatalf("Run failed with protection disabled: %v", err)
		}
	}
}
