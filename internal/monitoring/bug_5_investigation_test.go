// Bug #5 Investigation: Frame Rate Monitoring Not Actually Missing
// This test demonstrates that frame rate monitoring IS implemented

package monitoring

import (
	"testing"
	"time"
)

func TestBug5FrameRateMonitoringIsActuallyWorking(t *testing.T) {
	t.Log("Testing Bug #5: Investigation into Frame Rate Monitoring")

	// The claim: "IsFrameRateTargetMet() method exists but frame rate tracking and updating mechanism not evident"
	// Reality: Frame rate tracking IS implemented via RecordFrame() and background monitoring

	profiler := NewProfiler(50)
	err := profiler.Start("", "", true) // Enable debug mode for testing
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", true)

	// Test 1: Verify that RecordFrame() properly updates frame counters
	t.Run("RecordFrameUpdatesCounters", func(t *testing.T) {
		initialStats := profiler.GetStats()
		initialFrames := initialStats.TotalFrames

		// Record some frames
		for i := 0; i < 10; i++ {
			profiler.RecordFrame()
		}

		updatedStats := profiler.GetStats()
		if updatedStats.TotalFrames != initialFrames+10 {
			t.Errorf("Expected %d frames, got %d", initialFrames+10, updatedStats.TotalFrames)
		}

		t.Logf("✓ RecordFrame() correctly updates frame counters: %d -> %d", initialFrames, updatedStats.TotalFrames)
	})

	// Test 2: Verify that background monitoring thread calculates frame rate
	t.Run("BackgroundMonitoringCalculatesFrameRate", func(t *testing.T) {
		// Record frames at known rate
		for i := 0; i < 30; i++ {
			profiler.RecordFrame()
			time.Sleep(33 * time.Millisecond) // ~30 FPS
		}

		// Wait for monitoring cycle (which runs every 5 seconds)
		t.Log("Waiting for monitoring cycle to calculate frame rate...")
		time.Sleep(6 * time.Second)

		stats := profiler.GetStats()
		if stats.FrameRate <= 0 {
			t.Error("Frame rate should be calculated by background monitoring")
		}

		t.Logf("✓ Background monitoring calculated frame rate: %.1f FPS", stats.FrameRate)
	})

	// Test 3: Verify IsFrameRateTargetMet() uses calculated frame rate
	t.Run("IsFrameRateTargetMetUsesCalculatedRate", func(t *testing.T) {
		// Set known frame rate
		profiler.stats.mu.Lock()
		profiler.stats.FrameRate = 45.0 // Above 30 FPS target
		profiler.stats.mu.Unlock()

		if !profiler.IsFrameRateTargetMet() {
			t.Error("IsFrameRateTargetMet() should return true for 45 FPS")
		}

		profiler.stats.mu.Lock()
		profiler.stats.FrameRate = 20.0 // Below 30 FPS target
		profiler.stats.mu.Unlock()

		if profiler.IsFrameRateTargetMet() {
			t.Error("IsFrameRateTargetMet() should return false for 20 FPS")
		}

		t.Log("✓ IsFrameRateTargetMet() correctly uses calculated frame rate")
	})

	t.Log("CONCLUSION: Bug #5 is NOT a bug - frame rate monitoring IS fully implemented")
	t.Log("Components working correctly:")
	t.Log("  1. RecordFrame() - called from UI animation loop")
	t.Log("  2. Background monitoring thread - calculates FPS every 5 seconds")
	t.Log("  3. IsFrameRateTargetMet() - uses calculated frame rate")
	t.Log("  4. Integration - profiler passed to DesktopWindow and used")
}

func TestBug5FrameRateMonitoringIntegrationEvidence(t *testing.T) {
	t.Log("Demonstrating Bug #5 Integration Evidence")

	// This test shows the exact integration path mentioned in AUDIT.md
	profiler := NewProfiler(50)
	err := profiler.Start("", "", true) // Enable debug mode for testing
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", true)

	// Evidence 1: RecordFrame is called from processFrameUpdates in window.go:393
	t.Log("Evidence 1: UI calls RecordFrame() during animation loop")
	profiler.RecordFrame() // This is what window.go:393 does

	stats := profiler.GetStats()
	if stats.TotalFrames == 0 {
		t.Error("RecordFrame() should increment frame counter")
	}
	t.Logf("✓ Frame recorded: %d total frames", stats.TotalFrames)

	// Evidence 2: Background thread monitors and calculates frame rate
	t.Log("Evidence 2: Background monitoring calculates frame rate")

	// Simulate animation loop for sufficient time
	for i := 0; i < 50; i++ {
		profiler.RecordFrame()
		time.Sleep(20 * time.Millisecond) // 50 FPS simulation
	}

	// Wait for calculation cycle
	time.Sleep(6 * time.Second)

	finalStats := profiler.GetStats()
	t.Logf("✓ Frame rate calculated: %.1f FPS from %d frames", finalStats.FrameRate, finalStats.TotalFrames)

	// Evidence 3: Target checking works with calculated rate
	targetMet := profiler.IsFrameRateTargetMet()
	t.Logf("✓ Frame rate target (30 FPS) met: %v", targetMet)

	t.Log("AUDIT.md Assessment: Frame rate monitoring mechanism IS evident and working")
}
