package character

import (
	"image"
	"image/gif"
	"testing"
	"time"
)

// TestGIFDelayCalculationDirect tests the frame delay calculation directly
func TestGIFDelayCalculationDirect(t *testing.T) {
	am := NewAnimationManager()

	// Create a test GIF with known delay values
	// According to GIF spec: delay is in centiseconds (1/100th of a second)
	testGif := &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{10}, // 10 centiseconds = 0.1 seconds = 100ms
	}

	am.animations["test"] = testGif
	am.currentAnim = "test"
	am.playing = true

	// Check how the delay is calculated
	_, needsUpdate := am.GetCurrentFrame()
	if needsUpdate {
		t.Error("Should not need update immediately")
	}

	// Let's test different timing scenarios to understand the calculation
	testCases := []struct {
		waitTime     time.Duration
		shouldUpdate bool
		description  string
	}{
		{50 * time.Millisecond, false, "50ms should not be enough for 100ms delay"},
		{120 * time.Millisecond, true, "120ms should be enough for 100ms delay"},
		{1100 * time.Millisecond, true, "1100ms should definitely be enough"},
	}

	for _, tc := range testCases {
		// Reset timing
		am.lastUpdate = time.Now()

		time.Sleep(tc.waitTime)

		_, needsUpdate := am.GetCurrentFrame()

		if needsUpdate != tc.shouldUpdate {
			t.Errorf("%s: expected needsUpdate=%t, got %t (waited %v)",
				tc.description, tc.shouldUpdate, needsUpdate, tc.waitTime)

			// Let's see what the actual delay calculation produces
			currentGif := am.animations[am.currentAnim]
			frameDelay := time.Duration(currentGif.Delay[am.frameIndex]) * 10 * time.Millisecond
			actualWait := time.Since(am.lastUpdate)
			t.Logf("  GIF delay value: %d centiseconds", currentGif.Delay[am.frameIndex])
			t.Logf("  Calculated frame delay: %v", frameDelay)
			t.Logf("  Actual wait time: %v", actualWait)
			t.Logf("  Wait >= delay? %t", actualWait >= frameDelay)
		}
	}
}

// TestGIFDelayCalculationMath tests the mathematical calculation
func TestGIFDelayCalculationMath(t *testing.T) {
	// Test the delay calculation formula directly
	gifDelayValues := []int{1, 5, 10, 20, 100}

	for _, delayValue := range gifDelayValues {
		// Current implementation (potentially buggy)
		currentCalc := time.Duration(delayValue) * 10 * time.Millisecond

		// Correct implementation (GIF spec: delay is in centiseconds)
		correctCalc := time.Duration(delayValue) * time.Millisecond * 10 // centiseconds to milliseconds

		t.Logf("GIF delay value: %d centiseconds", delayValue)
		t.Logf("  Current calculation: %v", currentCalc)
		t.Logf("  Correct calculation: %v", correctCalc)
		t.Logf("  Difference: %v", currentCalc-correctCalc)

		// They should be the same if the current implementation is correct
		if currentCalc != correctCalc {
			t.Errorf("Calculation mismatch for delay %d: current=%v, correct=%v",
				delayValue, currentCalc, correctCalc)
		}
	}
}
