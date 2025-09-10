package character

import (
	"image"
	"image/gif"
	"testing"
	"time"
)

// TestGIFFrameRateCalculationBug demonstrates Bug #7: GIF animations play 10x slower than intended
func TestGIFFrameRateCalculationBug(t *testing.T) {
	// Create an animation manager
	am := NewAnimationManager()

	// Create a test GIF with specific timing
	// GIF delay values are in centiseconds (10ms units)
	// A delay of 10 means 100ms per frame (10 FPS)
	testGif := &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{10, 10}, // 10 centiseconds = 100ms per frame = 10 FPS
	}

	// Load the test animation directly (using internal access for testing)
	am.animations["test"] = testGif
	am.currentAnim = "test"
	am.playing = true

	// BUG: Current implementation treats delay as milliseconds and multiplies by 10
	// So delay=10 becomes 10 * 10ms = 100ms, then * 10 again = 1000ms (1 second per frame)
	// Expected: delay=10 centiseconds = 100ms per frame
	// Actual: 1000ms per frame (10x slower)

	// Test the timing calculation
	frame1, needsUpdate1 := am.GetCurrentFrame()
	if frame1 == nil {
		t.Fatal("Expected to get a frame")
	}

	// Should not need update immediately
	if needsUpdate1 {
		t.Error("Frame should not need immediate update")
	}

	// Wait 150ms - should be enough for one frame at 100ms/frame (10 FPS)
	time.Sleep(150 * time.Millisecond)

	frame2, needsUpdate2 := am.GetCurrentFrame()
	if frame2 == nil {
		t.Fatal("Expected to get a frame")
	}

	// BUG: With current implementation, 150ms is not enough (needs 1000ms)
	if !needsUpdate2 {
		t.Errorf("BUG: Frame should need update after 150ms with 100ms/frame timing, but current implementation requires 1000ms (10x slower)")
		t.Logf("This demonstrates that GIF animations play 10x slower than intended")
	}

	// Wait total 1100ms to demonstrate the bug
	time.Sleep(950 * time.Millisecond) // Total wait: 150 + 950 = 1100ms

	frame3, needsUpdate3 := am.GetCurrentFrame()
	if frame3 == nil {
		t.Fatal("Expected to get a frame")
	}

	// With the bug, 1100ms should be enough to trigger frame update
	if !needsUpdate3 {
		t.Error("After 1100ms, frame should definitely need update (this shows the 10x slowdown)")
	}
}

// TestGIFFrameRateCalculationFix tests the corrected behavior
func TestGIFFrameRateCalculationFix(t *testing.T) {
	// This test will pass once the bug is fixed
	am := NewAnimationManager()

	// Create a test GIF with 5 centiseconds delay = 50ms per frame = 20 FPS
	testGif := &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{5, 5}, // 5 centiseconds = 50ms per frame = 20 FPS
	}

	// Load the test animation directly
	am.animations["test"] = testGif
	am.currentAnim = "test"
	am.playing = true

	// With correct implementation, 60ms should be enough for frame update (50ms + margin)
	time.Sleep(60 * time.Millisecond)

	_, needsUpdate := am.GetCurrentFrame()
	if !needsUpdate {
		t.Error("FIXED: Frame should need update after 60ms with 50ms/frame timing")
	}
}
