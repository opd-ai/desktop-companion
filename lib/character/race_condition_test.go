package character

import (
	"sync"
	"testing"
	"time"
)

// TestConcurrentAccess tests concurrent access to character state
// This test demonstrates proper synchronization using the character's mutex
func TestConcurrentAccess(t *testing.T) {
	// Create a character struct with initial state
	char := &Character{
		currentState: "idle", // Initial state
	}

	// This will use proper synchronization with the character's mutex
	done := make(chan bool, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: writes to character state with proper mutex protection
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			char.mu.Lock()
			char.currentState = "talking" // Writing with mutex protection
			char.mu.Unlock()
			time.Sleep(1 * time.Microsecond)
		}
		done <- true
	}()

	// Goroutine 2: writes to character state with proper mutex protection
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			char.mu.Lock()
			char.currentState = "happy" // Writing with mutex protection
			char.mu.Unlock()
			time.Sleep(1 * time.Microsecond)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	wg.Wait()

	// Read the final state with proper synchronization
	char.mu.RLock()
	finalState := char.currentState
	char.mu.RUnlock()

	// This should pass with no race conditions
	if finalState != "talking" && finalState != "happy" {
		t.Errorf("Expected character state to be either talking or happy, got: %s", finalState)
	}
}
