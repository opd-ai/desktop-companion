package dialog

import (
	"sync"
	"testing"
	"time"
)

// TestDialogManagerConcurrency verifies that the DialogManager is safe for concurrent access
func TestDialogManagerConcurrency(t *testing.T) {
	manager := NewDialogManager(false)

	// Create backends
	simpleBackend := NewSimpleRandomBackend()
	markovBackend := NewMarkovChainBackend()

	// Initialize backends
	simpleBackend.Initialize([]byte(`{}`))
	markovBackend.Initialize([]byte(`{"chainOrder": 2, "trainingData": ["Hello", "World"]}`))

	// Number of concurrent goroutines
	numGoroutines := 20
	numOperations := 100

	var wg sync.WaitGroup

	// Test concurrent registration and configuration
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				// Concurrent operations that should be safe
				switch j % 6 {
				case 0:
					// Register backends
					manager.RegisterBackend("simple", simpleBackend)
				case 1:
					// Set default backend
					manager.SetDefaultBackend("simple")
				case 2:
					// Get registered backends
					manager.GetRegisteredBackends()
				case 3:
					// Get backend
					manager.GetBackend("simple")
				case 4:
					// Set fallback chain
					manager.SetFallbackChain([]string{"simple"})
				case 5:
					// Generate dialog
					context := DialogContext{
						Trigger:           "test",
						FallbackResponses: []string{"Hello"},
					}
					manager.GenerateDialog(context)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify the manager is still functional after concurrent access
	backends := manager.GetRegisteredBackends()
	if len(backends) == 0 {
		t.Error("Expected at least one registered backend after concurrent operations")
	}

	// Test dialog generation still works
	context := DialogContext{
		Trigger:           "final_test",
		FallbackResponses: []string{"Final test"},
	}

	response, err := manager.GenerateDialog(context)
	if err != nil {
		t.Errorf("Dialog generation failed after concurrent operations: %v", err)
	}

	if response.Text == "" {
		t.Error("Expected non-empty response after concurrent operations")
	}
}

// TestDialogManagerReadWritePatterns tests typical read/write patterns
func TestDialogManagerReadWritePatterns(t *testing.T) {
	manager := NewDialogManager(false)

	// Initialize with some backends
	simpleBackend := NewSimpleRandomBackend()
	simpleBackend.Initialize([]byte(`{}`))
	manager.RegisterBackend("simple", simpleBackend)
	manager.SetDefaultBackend("simple")

	numReaders := 10
	numWriters := 2
	duration := 200 * time.Millisecond

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start readers (should be able to run concurrently)
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					// Read operations
					manager.GetRegisteredBackends()
					manager.GetBackend("simple")
					context := DialogContext{
						Trigger:           "read_test",
						FallbackResponses: []string{"Read test"},
					}
					manager.GenerateDialog(context)
					time.Sleep(time.Millisecond)
				}
			}
		}()
	}

	// Start writers (will compete for write locks)
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					// Write operations
					backend := NewSimpleRandomBackend()
					backend.Initialize([]byte(`{}`))

					backendName := "dynamic_backend"
					manager.RegisterBackend(backendName, backend)
					manager.SetFallbackChain([]string{backendName})
					time.Sleep(5 * time.Millisecond)
				}
			}
		}(i)
	}

	// Let them run for a while
	time.Sleep(duration)
	close(done)
	wg.Wait()

	// Verify manager is still functional
	backends := manager.GetRegisteredBackends()
	if len(backends) == 0 {
		t.Error("Expected backends to be registered after read/write test")
	}
}
