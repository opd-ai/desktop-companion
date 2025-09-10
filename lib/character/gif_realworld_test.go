package character

import (
	"image"
	"image/gif"
	"testing"
	"time"
)

// TestGIFFrameRateRealWorld tests if animations play at expected speeds
func TestGIFFrameRateRealWorld(t *testing.T) {
	am := NewAnimationManager()

	// Create a test GIF that should play at 10 FPS (100ms per frame)
	// If bug exists, it would play at 1 FPS (1000ms per frame)
	fastGif := &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{2}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{10, 10, 10}, // 10 centiseconds each = 100ms = 10 FPS total
	}

	am.animations["fast"] = fastGif
	am.currentAnim = "fast"
	am.playing = true

	// Track frame updates over time
	startTime := time.Now()
	am.lastUpdate = startTime
	frameUpdates := 0
	maxTestTime := 350 * time.Millisecond // Should see 3 frame updates in 300ms at 10 FPS

	for time.Since(startTime) < maxTestTime {
		time.Sleep(10 * time.Millisecond) // Check every 10ms

		if am.Update() { // Returns true if frame was updated
			frameUpdates++
			t.Logf("Frame update #%d at %v", frameUpdates, time.Since(startTime))
		}
	}

	totalTime := time.Since(startTime)
	expectedUpdates := 3 // Should see ~3 updates in 350ms at 10 FPS (every 100ms)

	t.Logf("Total time: %v", totalTime)
	t.Logf("Frame updates: %d", frameUpdates)
	t.Logf("Expected updates: %d", expectedUpdates)

	// If the bug exists (10x slower), we'd see 0 updates in 350ms
	// If correct, we should see 3 updates
	if frameUpdates == 0 {
		t.Error("BUG CONFIRMED: No frame updates in 350ms suggests 10x slowdown")
	} else if frameUpdates >= 2 && frameUpdates <= 4 {
		t.Logf("CORRECT: Frame timing appears normal (%d updates in %v)", frameUpdates, totalTime)
	} else {
		t.Errorf("UNEXPECTED: Got %d frame updates, expected around %d", frameUpdates, expectedUpdates)
	}
}

// TestGIFFrameRateValidation tests if AUDIT.md bug report is accurate
func TestGIFFrameRateValidation(t *testing.T) {
	// Let's test the exact claim: "animations play 10x slower than intended"

	am := NewAnimationManager()

	// Create test GIF with 5 centiseconds delay = should be 50ms per frame = 20 FPS
	testGif := &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{5, 5}, // 5 centiseconds = 50ms per frame
	}

	am.animations["test"] = testGif
	am.currentAnim = "test"
	am.playing = true

	// Test if frame timing matches expectation
	startTime := time.Now()
	am.lastUpdate = startTime

	// Wait 60ms - should be enough for one frame at 50ms timing
	time.Sleep(60 * time.Millisecond)

	_, needsUpdate := am.GetCurrentFrame()
	elapsed := time.Since(startTime)

	t.Logf("Elapsed time: %v", elapsed)
	t.Logf("Needs update after 60ms: %t", needsUpdate)

	if !needsUpdate {
		// Check if it needs much longer (indicating 10x slowdown)
		time.Sleep(450 * time.Millisecond) // Total wait: 510ms
		_, needsUpdateLater := am.GetCurrentFrame()

		if needsUpdateLater {
			t.Error("BUG CONFIRMED: Needs 500ms+ for 50ms frame timing (10x slowdown)")
		} else {
			t.Error("UNEXPECTED: Frame timing appears broken in some other way")
		}
	} else {
		t.Log("CORRECT: Frame timing works as expected (no 10x slowdown)")
	}
}
