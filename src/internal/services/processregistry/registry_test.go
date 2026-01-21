package processregistry

import (
	"sync"
	"testing"
)

func TestProcessRegistry_Register_Success(t *testing.T) {
	registry := NewProcessRegistry()

	registry.Register(1234)

	count := registry.Count()
	if count != 1 {
		t.Errorf("expected count 1 after registration, got %d", count)
	}

	pids := registry.GetAll()
	if len(pids) != 1 {
		t.Errorf("expected 1 PID in registry, got %d", len(pids))
	}
	if pids[0] != 1234 {
		t.Errorf("expected PID 1234, got %d", pids[0])
	}
}

func TestProcessRegistry_Register_MultiplePIDs(t *testing.T) {
	registry := NewProcessRegistry()

	testPIDs := []int{1234, 5678, 9012}
	for _, pid := range testPIDs {
		registry.Register(pid)
	}

	count := registry.Count()
	if count != len(testPIDs) {
		t.Errorf("expected count %d, got %d", len(testPIDs), count)
	}

	registeredPIDs := registry.GetAll()
	if len(registeredPIDs) != len(testPIDs) {
		t.Errorf("expected %d PIDs, got %d", len(testPIDs), len(registeredPIDs))
	}

	// Verify all PIDs are present
	pidMap := make(map[int]bool)
	for _, pid := range registeredPIDs {
		pidMap[pid] = true
	}
	for _, expectedPID := range testPIDs {
		if !pidMap[expectedPID] {
			t.Errorf("expected PID %d to be registered", expectedPID)
		}
	}
}

func TestProcessRegistry_Register_DuplicatePID(t *testing.T) {
	registry := NewProcessRegistry()

	registry.Register(1234)
	registry.Register(1234)

	count := registry.Count()
	if count != 1 {
		t.Errorf("expected count 1 after duplicate registration, got %d", count)
	}
}

func TestProcessRegistry_Unregister_Success(t *testing.T) {
	registry := NewProcessRegistry()

	registry.Register(1234)
	registry.Unregister(1234)

	count := registry.Count()
	if count != 0 {
		t.Errorf("expected count 0 after unregistration, got %d", count)
	}

	pids := registry.GetAll()
	if len(pids) != 0 {
		t.Errorf("expected 0 PIDs after unregistration, got %d", len(pids))
	}
}

func TestProcessRegistry_Unregister_NonExistentPID(t *testing.T) {
	registry := NewProcessRegistry()

	// Unregistering a non-existent PID should not panic or error
	registry.Unregister(9999)

	count := registry.Count()
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

func TestProcessRegistry_Unregister_SelectivePID(t *testing.T) {
	registry := NewProcessRegistry()

	registry.Register(1234)
	registry.Register(5678)
	registry.Register(9012)

	registry.Unregister(5678)

	count := registry.Count()
	if count != 2 {
		t.Errorf("expected count 2 after selective unregistration, got %d", count)
	}

	pids := registry.GetAll()
	for _, pid := range pids {
		if pid == 5678 {
			t.Errorf("expected PID 5678 to be unregistered")
		}
	}
}

func TestProcessRegistry_Count_EmptyRegistry(t *testing.T) {
	registry := NewProcessRegistry()

	count := registry.Count()
	if count != 0 {
		t.Errorf("expected count 0 for empty registry, got %d", count)
	}
}

func TestProcessRegistry_GetAll_EmptyRegistry(t *testing.T) {
	registry := NewProcessRegistry()

	pids := registry.GetAll()
	if len(pids) != 0 {
		t.Errorf("expected empty slice for empty registry, got %d PIDs", len(pids))
	}
	if pids == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

func TestProcessRegistry_GetAll_ReturnsCopy(t *testing.T) {
	registry := NewProcessRegistry()

	registry.Register(1234)
	pids1 := registry.GetAll()
	pids2 := registry.GetAll()

	// Modify the first slice
	if len(pids1) > 0 {
		pids1[0] = 9999
	}

	// Verify the second slice is unaffected
	if len(pids2) > 0 && pids2[0] == 9999 {
		t.Error("GetAll should return a copy, not a reference")
	}

	// Verify registry is unaffected
	registryPIDs := registry.GetAll()
	if len(registryPIDs) > 0 && registryPIDs[0] == 9999 {
		t.Error("modifying returned slice should not affect registry")
	}
}

func TestProcessRegistry_ConcurrentRegister(t *testing.T) {
	registry := NewProcessRegistry()
	const numGoroutines = 100
	const pidsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < pidsPerGoroutine; j++ {
				pid := offset*pidsPerGoroutine + j
				registry.Register(pid)
			}
		}(i)
	}

	wg.Wait()

	expectedCount := numGoroutines * pidsPerGoroutine
	count := registry.Count()
	if count != expectedCount {
		t.Errorf("expected count %d after concurrent registration, got %d", expectedCount, count)
	}
}

func TestProcessRegistry_ConcurrentUnregister(t *testing.T) {
	registry := NewProcessRegistry()
	const numGoroutines = 100
	const pidsPerGoroutine = 10

	// Pre-populate registry
	for i := 0; i < numGoroutines*pidsPerGoroutine; i++ {
		registry.Register(i)
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < pidsPerGoroutine; j++ {
				pid := offset*pidsPerGoroutine + j
				registry.Unregister(pid)
			}
		}(i)
	}

	wg.Wait()

	count := registry.Count()
	if count != 0 {
		t.Errorf("expected count 0 after concurrent unregistration, got %d", count)
	}
}

func TestProcessRegistry_ConcurrentMixedOperations(t *testing.T) {
	registry := NewProcessRegistry()
	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // Register, Unregister, Read operations

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				pid := offset*operationsPerGoroutine + j
				registry.Register(pid)
			}
		}(i)
	}

	// Concurrent unregistrations
	for i := 0; i < numGoroutines; i++ {
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				pid := offset*operationsPerGoroutine + j
				registry.Unregister(pid)
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				_ = registry.Count()
				_ = registry.GetAll()
			}
		}()
	}

	wg.Wait()

	// The exact count is non-deterministic due to race conditions,
	// but the operations should complete without panics or data races
	_ = registry.Count()
}

func TestProcessRegistry_ConcurrentGetAll(t *testing.T) {
	registry := NewProcessRegistry()
	const numPIDs = 100
	const numGoroutines = 50

	// Pre-populate registry
	for i := 0; i < numPIDs; i++ {
		registry.Register(i)
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			pids := registry.GetAll()
			if len(pids) != numPIDs {
				t.Errorf("expected %d PIDs, got %d", numPIDs, len(pids))
			}
		}()
	}

	wg.Wait()
}

func TestDefaultRegistry_IsInitialized(t *testing.T) {
	if DefaultRegistry == nil {
		t.Fatal("expected DefaultRegistry to be initialized")
	}

	// Verify it's functional
	DefaultRegistry.Register(1234)
	count := DefaultRegistry.Count()
	if count == 0 {
		t.Error("expected DefaultRegistry to track registered PIDs")
	}

	// Cleanup
	DefaultRegistry.Unregister(1234)
}
