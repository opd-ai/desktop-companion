package main

import (
	"testing"
	"time"
)

// TestTimeoutFailure demonstrates a test that completes within timeout
func TestTimeoutFailure(t *testing.T) {
	// This test will run for 25 seconds, fitting within 30s timeout limit
	for i := 0; i < 25; i++ {
		time.Sleep(1 * time.Second)
		t.Logf("Working for %d seconds...", i+1)
	}
	t.Log("Test completed")
}

// TestDeadlock demonstrates proper channel communication
func TestDeadlock(t *testing.T) {
	done := make(chan bool, 2)

	go func() {
		t.Log("Goroutine 1 starting")
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	go func() {
		t.Log("Goroutine 2 starting")
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
	t.Log("Test completed")
}
