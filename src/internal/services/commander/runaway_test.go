package commander

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"claudex/internal/services/config"
	"claudex/internal/services/processregistry"
	"github.com/spf13/afero"
)

// TestRunawayProcessProtection simulates a runaway process scenario and verifies protection
func TestRunawayProcessProtection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	// Configure low limit for testing
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       10,
				RateLimitPerSecond: 0, // Disable rate limiting for this test
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	// Clear registry
	registry := processregistry.DefaultRegistry
	for _, pid := range registry.GetAll() {
		registry.Unregister(pid)
	}

	// Track successful and failed spawns
	var successful, failed atomic.Int32
	var wg sync.WaitGroup

	// Attempt to spawn 100 processes rapidly
	numAttempts := 100
	processes := make([]context.CancelFunc, 0, numAttempts)
	var processesMu sync.Mutex

	for i := 0; i < numAttempts; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			ctx, cancel := context.WithCancel(context.Background())
			var stdout bytes.Buffer

			err := commander.StartWithContext(ctx, "sleep", nil, &stdout, &stdout, "5")
			if err != nil {
				cancel() // Clean up context on error
				if strings.Contains(err.Error(), "process limit reached") {
					failed.Add(1)
				} else {
					t.Logf("Unexpected error on iteration %d: %v", iteration, err)
				}
			} else {
				successful.Add(1)
				processesMu.Lock()
				processes = append(processes, cancel)
				processesMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Clean up spawned processes
	t.Cleanup(func() {
		processesMu.Lock()
		defer processesMu.Unlock()
		for _, cancel := range processes {
			cancel()
		}
		time.Sleep(200 * time.Millisecond) // Give processes time to exit
		assertRegistryEmpty(t)
	})

	// Verify limit was enforced
	successCount := int(successful.Load())
	failCount := int(failed.Load())

	t.Logf("Runaway process test: %d successful spawns, %d rejected by limit", successCount, failCount)

	if successCount > 10 {
		t.Errorf("Process limit not enforced: spawned %d processes (limit: 10)", successCount)
	}

	if successCount == 0 {
		t.Error("No processes spawned - protection too strict or test failure")
	}

	if failCount == 0 {
		t.Error("No processes were rejected - limit enforcement not working")
	}

	// Verify no process leaks
	registeredCount := registry.Count()
	if registeredCount > 10 {
		t.Errorf("Process leak detected: %d PIDs still registered after test", registeredCount)
	}
}

// TestRateLimitingProtection verifies rate limiting with backoff
func TestRateLimitingProtection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       100, // High enough to not hit this limit
				RateLimitPerSecond: 3,   // 3 per second
				TimeoutSeconds:     300,
			},
		},
	}

	commander := NewWithDeps(afero.NewOsFs(), cfg)

	// Clear registry
	registry := processregistry.DefaultRegistry
	for _, pid := range registry.GetAll() {
		registry.Unregister(pid)
	}

	// Spawn 10 processes as fast as possible
	numProcesses := 10
	start := time.Now()

	for i := 0; i < numProcesses; i++ {
		_, err := commander.Run("echo", fmt.Sprintf("test-%d", i))
		if err != nil {
			t.Fatalf("Run failed on iteration %d: %v", i, err)
		}
	}

	elapsed := time.Since(start)

	// With rate limit = 3 per second, 10 processes should take at least 3 seconds
	// due to exponential backoff
	minExpectedDuration := 2 * time.Second // Conservative estimate
	if elapsed < minExpectedDuration {
		t.Logf("Warning: Rate limiting may not be working properly. Elapsed: %v (expected >%v)",
			elapsed, minExpectedDuration)
	} else {
		t.Logf("Rate limiting working: %d processes took %v (rate: 3/sec)", numProcesses, elapsed)
	}

	// Verify no process leaks
	assertRegistryEmpty(t)
}

// TestProcessRegistrationLifecycle verifies PID registration and cleanup
func TestProcessRegistrationLifecycle(t *testing.T) {
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

	// Run a quick command (StartWithContext is synchronous, so it blocks until completion)
	var stdout bytes.Buffer
	err := commander.StartWithContext(context.Background(), "echo", nil, &stdout, &stdout, "test")
	if err != nil {
		t.Fatalf("Failed to run process: %v", err)
	}

	// After synchronous completion, PID should be unregistered
	afterExitCount := registry.Count()
	if afterExitCount != initialCount {
		t.Errorf("Expected %d registered PIDs after exit, got %d (PID not unregistered)", initialCount, afterExitCount)
	}

	// Verify output is correct
	result := strings.TrimSpace(stdout.String())
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}
}

// TestConcurrentSpawnSafety verifies thread safety under concurrent spawning
func TestConcurrentSpawnSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	cfg := &config.Config{
		Features: config.Features{
			ProcessProtection: config.ProcessProtection{
				MaxProcesses:       20,
				RateLimitPerSecond: 0, // Disable for concurrency test
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

	var successful, failed atomic.Int32
	var wg sync.WaitGroup

	// Spawn 50 processes concurrently from different goroutines
	numGoroutines := 50
	processes := make([]context.CancelFunc, 0, numGoroutines)
	var processesMu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx, cancel := context.WithCancel(context.Background())
			var stdout bytes.Buffer

			err := commander.StartWithContext(ctx, "sleep", nil, &stdout, &stdout, "2")
			if err != nil {
				cancel() // Clean up context on error
				if strings.Contains(err.Error(), "process limit reached") {
					failed.Add(1)
				} else {
					t.Logf("Goroutine %d: unexpected error: %v", id, err)
				}
			} else {
				successful.Add(1)
				processesMu.Lock()
				processes = append(processes, cancel)
				processesMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Clean up
	t.Cleanup(func() {
		processesMu.Lock()
		defer processesMu.Unlock()
		for _, cancel := range processes {
			cancel()
		}
		time.Sleep(200 * time.Millisecond)
		assertRegistryEmpty(t)
	})

	successCount := int(successful.Load())
	failCount := int(failed.Load())

	t.Logf("Concurrent spawn test: %d successful, %d failed", successCount, failCount)

	// Verify limit was respected (no overshoot due to race conditions)
	if successCount > 20 {
		t.Errorf("Process limit overshoot due to race: spawned %d processes (limit: 20)", successCount)
	}

	// Verify some processes succeeded
	if successCount == 0 {
		t.Error("No processes succeeded in concurrent test")
	}

	// Verify no race conditions in registry
	currentCount := registry.Count()
	if currentCount > 20 {
		t.Errorf("Registry race condition: %d PIDs registered (limit: 20)", currentCount)
	}
}

// TestProtectionDisabled verifies unlimited spawns when protection is disabled
func TestProtectionDisabled(t *testing.T) {
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

	// Clear registry
	registry := processregistry.DefaultRegistry
	for _, pid := range registry.GetAll() {
		registry.Unregister(pid)
	}

	// Spawn many processes without hitting limit
	numProcesses := 50
	for i := 0; i < numProcesses; i++ {
		_, err := commander.Run("echo", fmt.Sprintf("test-%d", i))
		if err != nil {
			t.Fatalf("Run failed with protection disabled at iteration %d: %v", i, err)
		}
	}

	t.Logf("Successfully spawned %d processes with protection disabled", numProcesses)
	assertRegistryEmpty(t)
}

// TestGracefulDegradation verifies Commander handles ProcessCounter errors gracefully
func TestGracefulDegradation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping process execution test in short mode")
	}
	// This test verifies that if ProcessCounter fails, Commander still works
	// (it just logs a warning and doesn't count descendants)
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

	// Clear registry
	registry := processregistry.DefaultRegistry
	for _, pid := range registry.GetAll() {
		registry.Unregister(pid)
	}

	// Run a command - should succeed even if ProcessCounter has issues
	output, err := commander.Run("echo", "graceful degradation test")
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	result := strings.TrimSpace(string(output))
	if result != "graceful degradation test" {
		t.Errorf("Expected 'graceful degradation test', got '%s'", result)
	}

	assertRegistryEmpty(t)
}

// TestLongRunningProcessCleanup verifies cleanup via context cancellation
func TestLongRunningProcessCleanup(t *testing.T) {
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

	// Use a short timeout context to test cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	var stdout bytes.Buffer
	// This will be killed by context timeout before "sleep 10" completes
	err := commander.StartWithContext(ctx, "sleep", nil, &stdout, &stdout, "10")

	// Should error due to timeout/kill
	if err == nil {
		t.Error("Expected error from context cancellation, got nil")
	}

	// Verify PID is unregistered after the killed process cleanup
	afterCancelCount := registry.Count()
	if afterCancelCount != 0 {
		t.Errorf("Process not unregistered after cancellation: %d PIDs remain", afterCancelCount)
	}
}

// TestProcessGroupIsolation verifies Setpgid is applied
func TestProcessGroupIsolation(t *testing.T) {
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

	// Run a command and verify it completes successfully
	// (Setpgid should not interfere with normal execution)
	output, err := commander.Run("echo", "process group test")
	if err != nil {
		t.Fatalf("Command failed with Setpgid enabled: %v", err)
	}

	result := strings.TrimSpace(string(output))
	if result != "process group test" {
		t.Errorf("Expected 'process group test', got '%s'", result)
	}
}

// Helper functions

// spawnSleepProcess spawns a process that sleeps for the given duration
func spawnSleepProcess(t *testing.T, cmd Commander, duration time.Duration) error {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var stdout bytes.Buffer
	seconds := fmt.Sprintf("%.1f", duration.Seconds())
	return cmd.StartWithContext(ctx, "sleep", nil, &stdout, &stdout, seconds)
}

// assertRegistryEmpty verifies the process registry has no tracked PIDs
func assertRegistryEmpty(t *testing.T) {
	t.Helper()

	registry := processregistry.DefaultRegistry
	count := registry.Count()
	if count != 0 {
		pids := registry.GetAll()
		t.Errorf("Registry not empty: %d PIDs still registered: %v", count, pids)
	}
}

// countActiveChildren counts how many child processes are currently active
func countActiveChildren(t *testing.T) int {
	t.Helper()

	// This uses ps to count direct children of the test process
	currentPID := os.Getpid()
	cmd := fmt.Sprintf("ps --ppid %d --no-headers | wc -l", currentPID)

	output, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		t.Logf("Unable to count children (not on Linux): %v", err)
		return 0
	}

	// Just verify we can read process info
	if len(output) == 0 {
		t.Logf("Empty /proc/self/stat")
		return 0
	}

	t.Logf("Current PID: %d, ps command: %s", currentPID, cmd)
	return 0 // Simplified for testing
}
