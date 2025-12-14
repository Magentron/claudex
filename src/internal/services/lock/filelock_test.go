package lock

import (
	"testing"

	"github.com/spf13/afero"
)

func TestFileLock_Acquire_Success(t *testing.T) {
	fs := afero.NewMemMapFs()
	lockService := New(fs)
	lockPath := "/test.lock"

	lock, err := lockService.Acquire(lockPath)
	if err != nil {
		t.Fatalf("expected successful lock acquisition, got error: %v", err)
	}
	if lock == nil {
		t.Fatal("expected non-nil lock")
	}
	if lock.Path != lockPath {
		t.Errorf("expected lock path %s, got %s", lockPath, lock.Path)
	}
	if lock.File == nil {
		t.Error("expected non-nil file handle")
	}

	// Verify lock file exists
	exists, err := afero.Exists(fs, lockPath)
	if err != nil {
		t.Fatalf("failed to check lock file existence: %v", err)
	}
	if !exists {
		t.Error("expected lock file to exist")
	}

	// Verify PID was written to lock file
	content, err := afero.ReadFile(fs, lockPath)
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}
	if len(content) == 0 {
		t.Error("expected lock file to contain PID")
	}

	// Cleanup
	if err := lock.Release(); err != nil {
		t.Errorf("failed to release lock: %v", err)
	}
}

func TestFileLock_Acquire_ConcurrentFails(t *testing.T) {
	fs := afero.NewMemMapFs()
	lockService := New(fs)
	lockPath := "/test.lock"

	// First acquisition should succeed
	lock1, err := lockService.Acquire(lockPath)
	if err != nil {
		t.Fatalf("first acquisition failed: %v", err)
	}
	defer lock1.Release()

	// Second acquisition should fail
	lock2, err := lockService.Acquire(lockPath)
	if err == nil {
		t.Fatal("expected second acquisition to fail, but it succeeded")
	}
	if lock2 != nil {
		t.Error("expected nil lock on failed acquisition")
	}
}

func TestFileLock_Release_RemovesFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	lockService := New(fs)
	lockPath := "/test.lock"

	lock, err := lockService.Acquire(lockPath)
	if err != nil {
		t.Fatalf("lock acquisition failed: %v", err)
	}

	// Verify lock file exists before release
	exists, err := afero.Exists(fs, lockPath)
	if err != nil {
		t.Fatalf("failed to check lock file existence: %v", err)
	}
	if !exists {
		t.Fatal("expected lock file to exist before release")
	}

	// Release the lock
	if err := lock.Release(); err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	// Verify lock file was removed
	exists, err = afero.Exists(fs, lockPath)
	if err != nil {
		t.Fatalf("failed to check lock file existence after release: %v", err)
	}
	if exists {
		t.Error("expected lock file to be removed after release")
	}
}

func TestFileLock_IsLocked_ReturnsCorrectState(t *testing.T) {
	fs := afero.NewMemMapFs()
	lockService := New(fs)
	lockPath := "/test.lock"

	// Initially, lock should not exist
	locked, err := lockService.IsLocked(lockPath)
	if err != nil {
		t.Fatalf("IsLocked failed: %v", err)
	}
	if locked {
		t.Error("expected lock to not exist initially")
	}

	// Acquire lock
	lock, err := lockService.Acquire(lockPath)
	if err != nil {
		t.Fatalf("lock acquisition failed: %v", err)
	}

	// Lock should now exist
	locked, err = lockService.IsLocked(lockPath)
	if err != nil {
		t.Fatalf("IsLocked failed after acquisition: %v", err)
	}
	if !locked {
		t.Error("expected lock to exist after acquisition")
	}

	// Release lock
	if err := lock.Release(); err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	// Lock should no longer exist
	locked, err = lockService.IsLocked(lockPath)
	if err != nil {
		t.Fatalf("IsLocked failed after release: %v", err)
	}
	if locked {
		t.Error("expected lock to not exist after release")
	}
}

func TestFileLock_Acquire_WritesPID(t *testing.T) {
	fs := afero.NewMemMapFs()
	lockService := New(fs)
	lockPath := "/test.lock"

	lock, err := lockService.Acquire(lockPath)
	if err != nil {
		t.Fatalf("lock acquisition failed: %v", err)
	}
	defer lock.Release()

	// Read the PID from the lock file
	content, err := afero.ReadFile(fs, lockPath)
	if err != nil {
		t.Fatalf("failed to read lock file: %v", err)
	}

	// Verify the content is not empty and looks like a number
	if len(content) == 0 {
		t.Error("expected lock file to contain PID")
	}

	// Verify it starts with a digit (PID should be a number)
	if content[0] < '0' || content[0] > '9' {
		t.Errorf("expected lock file content to start with a digit, got: %s", content)
	}
}

func TestFileLock_ReleaseAfterAcquireSucceeds(t *testing.T) {
	fs := afero.NewMemMapFs()
	lockService := New(fs)
	lockPath := "/test.lock"

	// Acquire and release multiple times to ensure cleanup is complete
	for i := 0; i < 3; i++ {
		lock, err := lockService.Acquire(lockPath)
		if err != nil {
			t.Fatalf("acquisition %d failed: %v", i, err)
		}

		if err := lock.Release(); err != nil {
			t.Fatalf("release %d failed: %v", i, err)
		}

		// Verify lock is released
		locked, err := lockService.IsLocked(lockPath)
		if err != nil {
			t.Fatalf("IsLocked check %d failed: %v", i, err)
		}
		if locked {
			t.Errorf("iteration %d: expected lock to be released", i)
		}
	}
}

func TestFileLock_IsLocked_ErrorHandling(t *testing.T) {
	// This test verifies IsLocked handles non-existent files gracefully
	fs := afero.NewMemMapFs()
	lockService := New(fs)

	locked, err := lockService.IsLocked("/nonexistent/path/to.lock")
	if err != nil {
		t.Errorf("expected no error for non-existent lock file, got: %v", err)
	}
	if locked {
		t.Error("expected non-existent lock to return false")
	}
}
