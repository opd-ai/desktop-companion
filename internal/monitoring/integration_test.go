package monitoring

import (
	"testing"
)

// TestProfilerIntegration tests basic profiler functionality
func TestProfilerIntegration(t *testing.T) {
	profiler := NewProfiler(50, 10)

	err := profiler.Start("", "", false)
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Record some metrics
	profiler.RecordStartupComplete()
	profiler.RecordFrame()

	stats := profiler.GetStats()
	if stats.TotalFrames != 1 {
		t.Errorf("Expected 1 frame, got %d", stats.TotalFrames)
	}

	if stats.StartupDuration == 0 {
		t.Error("Startup duration should be recorded")
	}
}
