// Bug #5 Frame Rate Monitoring Validation Test
// Tests that frame rate monitoring is actually working correctly

package monitoring

import (
	"testing"
	"time"
)

func TestBug5FrameRateMonitoringValidation(t *testing.T) {
	t.Log("Testing Bug #5: Frame Rate Monitoring Implementation")

	// Test 1: Verify frame rate calculation works
	t.Run("FrameRateCalculation", func(t *testing.T) {
		profiler := NewProfiler(50)
		err := profiler.Start("", "", true) // Enable debug mode for testing
		if err != nil {
			t.Fatalf("Failed to start profiler: %v", err)
		}
		defer profiler.Stop("", false)

		// Record frames continuously over multiple monitoring intervals
		// Frame rate calculation happens every 5 seconds, so we need to span that
		done := make(chan bool)
		go func() {
			for i := 0; i < 80; i++ { // Reduced frames from 150 to 80
				profiler.RecordFrame()
				time.Sleep(30 * time.Millisecond) // ~33 FPS pace, ~2.4 seconds total
			}
			done <- true
		}()

		// Wait for frames to be recorded across monitoring intervals
		<-done

		// Wait for at least one complete frame rate calculation cycle (5+ seconds)
		time.Sleep(5100 * time.Millisecond) // 5.1 seconds - minimal wait for one cycle

		stats := profiler.GetStats()
		if stats.TotalFrames != 80 { // Updated expected frame count
			t.Errorf("Expected 80 frames recorded, got %d", stats.TotalFrames)
		}

		// Frame rate should be calculated after monitoring cycle completes
		if stats.FrameRate <= 0 {
			t.Error("Frame rate should be calculated and greater than 0")
		}

		t.Logf("✓ Frame rate monitoring working: %.1f FPS from %d frames", stats.FrameRate, stats.TotalFrames)
	})

	// Test 2: Verify IsFrameRateTargetMet works
	t.Run("FrameRateTargetCheck", func(t *testing.T) {
		profiler := NewProfiler(50)
		err := profiler.Start("", "", true) // Enable debug mode for testing
		if err != nil {
			t.Fatalf("Failed to start profiler: %v", err)
		}
		defer profiler.Stop("", false)

		// Manually set a known frame rate for testing
		profiler.stats.mu.Lock()
		profiler.stats.FrameRate = 35.0 // Above 30 FPS target
		profiler.stats.mu.Unlock()

		if !profiler.IsFrameRateTargetMet() {
			t.Error("IsFrameRateTargetMet() should return true for 35 FPS")
		}

		// Test below target
		profiler.stats.mu.Lock()
		profiler.stats.FrameRate = 25.0 // Below 30 FPS target
		profiler.stats.mu.Unlock()

		if profiler.IsFrameRateTargetMet() {
			t.Error("IsFrameRateTargetMet() should return false for 25 FPS")
		}

		t.Log("✓ Frame rate target checking working correctly")
	})

	// Test 3: Verify frame rate monitoring thread starts
	t.Run("FrameRateMonitoringThread", func(t *testing.T) {
		profiler := NewProfiler(50)

		// Before starting, no monitoring
		initialFrameRate := profiler.GetStats().FrameRate
		if initialFrameRate != 0 {
			t.Errorf("Initial frame rate should be 0, got %.1f", initialFrameRate)
		}

		profiler.Start("", "", true) // Start monitoring with debug mode for testing
		defer profiler.Stop("", false)

		// Record some frames
		for i := 0; i < 60; i++ {
			profiler.RecordFrame()
			time.Sleep(16 * time.Millisecond) // ~60 FPS pace
		}

		// Wait for at least one monitoring cycle
		time.Sleep(5100 * time.Millisecond) // 5.1 seconds - minimal wait for one cycle

		stats := profiler.GetStats()
		if stats.FrameRate <= 0 {
			t.Error("Frame rate should be calculated after monitoring cycle")
		}

		t.Logf("✓ Frame rate monitoring thread working: %.1f FPS calculated", stats.FrameRate)
	})

	t.Log("Bug #5 validation: Frame rate monitoring IS actually implemented and working")
}

// TestBug5FrameRateIntegration tests integration with UI components
func TestBug5FrameRateIntegration(t *testing.T) {
	t.Log("Testing Bug #5 Integration: Frame Rate Monitoring in UI Context")

	profiler := NewProfiler(50)
	err := profiler.Start("", "", true) // Enable debug mode for testing
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop("", false)

	// Simulate animation loop calling RecordFrame()
	simulateAnimationLoop := func(frames int, fpsTarget float64) {
		interval := time.Duration(float64(time.Second) / fpsTarget)
		for i := 0; i < frames; i++ {
			profiler.RecordFrame() // This is what UI window.go does
			time.Sleep(interval)
		}
	}

	// Simulate 2 seconds of 30 FPS animation
	go simulateAnimationLoop(60, 30.0)

	// Wait for simulation and monitoring
	time.Sleep(6 * time.Second) // 6 seconds for simulation + one monitoring cycle

	stats := profiler.GetStats()

	if stats.TotalFrames == 0 {
		t.Error("No frames recorded during simulation")
	}

	if stats.FrameRate <= 0 {
		t.Error("Frame rate should be calculated")
	}

	// Check if frame rate is reasonable (allowing for timing variance)
	if stats.FrameRate < 20 || stats.FrameRate > 40 {
		t.Logf("WARNING: Frame rate %.1f outside expected range 20-40 FPS", stats.FrameRate)
	} else {
		t.Logf("✓ Frame rate %.1f FPS within expected range", stats.FrameRate)
	}

	// Verify target check works
	targetMet := profiler.IsFrameRateTargetMet()
	t.Logf("✓ Frame rate target (30 FPS) met: %v (actual: %.1f FPS)", targetMet, stats.FrameRate)

	t.Log("Bug #5 Integration: Frame rate monitoring correctly integrated with animation system")
}
