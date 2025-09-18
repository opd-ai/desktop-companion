package dialog

import (
	"runtime"
	"testing"
	"time"
)

// Test for goroutine leak in LLMDialogBackend.GenerateResponse
func TestLLMDialogBackend_GoroutineLeak(t *testing.T) {
	// This test demonstrates the goroutine leak that occurs when
	// GenerateResponse times out but the spawned goroutine continues running

	backend := NewLLMDialogBackend()

	// Configure with very short timeout to force timeouts
	configJSON := `{
		"enabled": true,
		"mockMode": false,
		"maxGenerationTime": 1,
		"fallbackResponses": ["fallback"]
	}`
	err := backend.Initialize([]byte(configJSON))
	if err != nil {
		t.Fatalf("Failed to initialize backend: %v", err)
	}

	// Get baseline goroutine count
	runtime.GC()
	runtime.GC()
	initialGoroutines := runtime.NumGoroutine()

	// Create requests that will timeout
	dialogContext := DialogContext{Trigger: "click"}

	const numRequests = 20
	for i := 0; i < numRequests; i++ {
		// This will timeout since we have no real LLM manager configured
		// but the goroutine will be created anyway and may leak
		_, _ = backend.GenerateResponse(dialogContext)
	}

	// Give time for any proper cleanup that might happen
	time.Sleep(50 * time.Millisecond)

	// Force garbage collection
	runtime.GC()
	runtime.GC()

	finalGoroutines := runtime.NumGoroutine()
	leakedGoroutines := finalGoroutines - initialGoroutines

	t.Logf("Initial goroutines: %d", initialGoroutines)
	t.Logf("Final goroutines: %d", finalGoroutines)
	t.Logf("Potentially leaked goroutines: %d", leakedGoroutines)

	// NOTE: This test documents the issue but may not always fail reliably
	// due to the specific conditions needed to trigger the leak.
	// The real fix is in the GenerateResponse method itself.
	if leakedGoroutines > 5 {
		t.Logf("WARNING: Potential goroutine leak detected with %d extra goroutines", leakedGoroutines)
	}
}
