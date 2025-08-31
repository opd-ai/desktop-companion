package main

import (
	"desktop-companion/internal/monitoring"
	"testing"
	"time"
)

// TestProfilerProductionOverhead tests that profiler doesn't create overhead when profiling is disabled
func TestProfilerProductionOverhead(t *testing.T) {
	// Create profiler but don't enable any file-based profiling (production scenario)
	profiler := monitoring.NewProfiler(50)

	// This represents how main.go uses the profiler - always calls Start()
	// even when no profiling files are specified
	err := profiler.Start("", "", false) // No mem profile, no CPU profile, no debug
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Wait a short time to see if monitoring goroutines are running
	time.Sleep(2 * time.Second)

	// Check if stats are being collected (indicating monitoring overhead)
	stats := profiler.GetStats()

	// BUG: The profiler collects stats even when no profiling is requested
	// In production, this creates unnecessary overhead
	if stats.CurrentMemoryMB > 0 {
		t.Errorf("Profiler is collecting memory stats when no profiling was requested - this creates production overhead")
		t.Logf("Memory stats being collected: %.2f MB", stats.CurrentMemoryMB)
	}

	// BUG: Frame rate monitoring also runs continuously
	if stats.TotalFrames > 0 {
		t.Errorf("Profiler is tracking frames when no profiling was requested - this creates production overhead")
		t.Logf("Frame tracking active: %d frames", stats.TotalFrames)
	}
}

// TestProfilerShouldOnlyMonitorWhenProfilePathsProvided tests the desired behavior
func TestProfilerShouldOnlyMonitorWhenProfilePathsProvided(t *testing.T) {
	// When profiling paths are provided, monitoring should be active
	profiler := monitoring.NewProfiler(50)

	err := profiler.Start("/tmp/test_mem.prof", "/tmp/test_cpu.prof", false)
	if err != nil {
		t.Fatalf("Failed to start profiler with profiling files: %v", err)
	}
	defer profiler.Stop("/tmp/test_mem.prof", false)

	time.Sleep(2 * time.Second)

	stats := profiler.GetStats()

	// When profiling is explicitly requested, stats collection is expected
	if stats.CurrentMemoryMB == 0 {
		t.Error("Profiler should collect stats when profiling paths are provided")
	}
}
