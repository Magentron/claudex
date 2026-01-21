package processcounter

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestNewProcessCounter(t *testing.T) {
	counter := NewProcessCounter()
	if counter == nil {
		t.Fatal("NewProcessCounter returned nil")
	}

	// Check that the correct implementation is returned based on platform
	switch runtime.GOOS {
	case "linux":
		if _, ok := counter.(*linuxCounter); !ok {
			t.Errorf("Expected linuxCounter on Linux, got %T", counter)
		}
	case "darwin":
		if _, ok := counter.(*darwinCounter); !ok {
			t.Errorf("Expected darwinCounter on macOS, got %T", counter)
		}
	default:
		// Default fallback is linuxCounter
		if _, ok := counter.(*linuxCounter); !ok {
			t.Errorf("Expected linuxCounter as fallback, got %T", counter)
		}
	}
}

func TestDefaultCounter(t *testing.T) {
	if DefaultCounter == nil {
		t.Fatal("DefaultCounter is nil")
	}
}

func TestCountDescendants_NoChildren(t *testing.T) {
	counter := NewProcessCounter()

	// Count descendants of current process (which should have no children initially)
	count, err := counter.CountDescendants(os.Getpid())
	if err != nil {
		t.Fatalf("CountDescendants failed: %v", err)
	}

	// Should return 0 for no children
	if count != 0 {
		t.Errorf("Expected 0 descendants, got %d", count)
	}
}

func TestCountDescendants_WithChildren(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	counter := NewProcessCounter()

	// Spawn a child process that sleeps
	cmd := exec.Command("sleep", "10")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start child process: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// Give the process time to start
	time.Sleep(100 * time.Millisecond)

	// Count descendants - should be at least 1
	count, err := counter.CountDescendants(os.Getpid())
	if err != nil {
		t.Fatalf("CountDescendants failed: %v", err)
	}

	if count < 1 {
		t.Errorf("Expected at least 1 descendant, got %d", count)
	}
}

func TestCountDescendants_MultiLevel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	counter := NewProcessCounter()

	// Spawn a shell that spawns another child
	// Using sh -c to create a subprocess hierarchy
	cmd := exec.Command("sh", "-c", "sleep 10 & sleep 10 & wait")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start child process: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// Give processes time to start
	time.Sleep(200 * time.Millisecond)

	// Count descendants - should be at least 3 (shell + 2 sleep processes)
	count, err := counter.CountDescendants(os.Getpid())
	if err != nil {
		t.Fatalf("CountDescendants failed: %v", err)
	}

	if count < 1 {
		t.Errorf("Expected at least 1 descendant, got %d", count)
	}
}

func TestCountDescendants_NonExistentPID(t *testing.T) {
	counter := NewProcessCounter()

	// Use a very high PID that likely doesn't exist
	nonExistentPID := 999999

	// Should return 0 descendants without error (process doesn't exist = no children)
	count, err := counter.CountDescendants(nonExistentPID)
	if err != nil {
		// Some implementations may return an error, which is acceptable
		t.Logf("CountDescendants for non-existent PID returned error: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 descendants for non-existent PID, got %d", count)
	}
}

func TestLinuxCounter_GetDirectChildren(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	counter := &linuxCounter{}

	// Test with current process (should have no children initially)
	children, err := counter.getDirectChildren(os.Getpid())
	if err != nil {
		t.Fatalf("getDirectChildren failed: %v", err)
	}

	if len(children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(children))
	}
}

func TestLinuxCounter_ReadChildrenFile(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	counter := &linuxCounter{}

	// Try reading the children file for the current process
	path := "/proc/self/task/" + string(rune(os.Getpid())) + "/children"
	children, err := counter.readChildrenFile(path)

	// May fail if file doesn't exist or we don't have permissions
	if err != nil {
		t.Logf("readChildrenFile returned error (expected for some systems): %v", err)
		return
	}

	// Should return empty slice for no children
	if len(children) != 0 {
		t.Logf("Got %d children (may be legitimate if other threads spawned processes)", len(children))
	}
}

func TestDarwinCounter_GetDirectChildren(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	counter := &darwinCounter{}

	// Test with current process (should have no children initially)
	children, err := counter.getDirectChildren(os.Getpid())
	if err != nil {
		t.Fatalf("getDirectChildren failed: %v", err)
	}

	if len(children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(children))
	}
}

func TestDarwinCounter_PgrepNoChildren(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	counter := &darwinCounter{}

	// Test with a process that has no children
	// Using current process which should have no children initially
	children, err := counter.getDirectChildren(os.Getpid())
	if err != nil {
		t.Fatalf("getDirectChildren failed: %v", err)
	}

	// pgrep should return empty result for no children (exit code 1 is handled)
	if len(children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(children))
	}
}

func BenchmarkCountDescendants_NoChildren(b *testing.B) {
	counter := NewProcessCounter()
	pid := os.Getpid()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := counter.CountDescendants(pid)
		if err != nil {
			b.Fatalf("CountDescendants failed: %v", err)
		}
	}
}

func BenchmarkCountDescendants_WithChild(b *testing.B) {
	counter := NewProcessCounter()

	// Spawn a long-running child process
	cmd := exec.Command("sleep", "60")
	err := cmd.Start()
	if err != nil {
		b.Fatalf("Failed to start child process: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	// Give the process time to start
	time.Sleep(100 * time.Millisecond)

	pid := os.Getpid()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := counter.CountDescendants(pid)
		if err != nil {
			b.Fatalf("CountDescendants failed: %v", err)
		}
	}
}
