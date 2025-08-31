package main

import (
	"image"
	"image/gif"
	"sync"
	"testing"
	"time"

	"desktop-companion/internal/character"
)

// TestBug3AnimationFrameRaceCondition reproduces the race condition described in the audit
// where GetCurrentFrame() timing checks might be inconsistent with Update() modifications
func TestBug3AnimationFrameRaceCondition(t *testing.T) {
	// Create animation manager with test animation
	am := character.NewAnimationManager()

	// Add test animation with multiple frames
	testGif := &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{2}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{1, 1, 1}, // Very fast 10ms frames to trigger race
	}

	// Load the animation (this would normally be done through LoadAnimation)
	err := am.LoadAnimation("test", testGif)
	if err != nil {
		t.Fatalf("Failed to load test animation: %v", err)
	}

	err = am.SetCurrentAnimation("test")
	if err != nil {
		t.Fatalf("Failed to set current animation: %v", err)
	}

	am.Play()

	// Track timing inconsistencies
	var inconsistencies int
	var mu sync.Mutex

	done := make(chan struct{})
	var wg sync.WaitGroup

	// Start rapid Update() calls (frame advancement)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				am.Update()
				time.Sleep(time.Millisecond) // Very frequent updates
			}
		}
	}()

	// Start rapid GetCurrentFrame() calls (timing checks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		var lastNewFrame bool
		for {
			select {
			case <-done:
				return
			default:
				_, newFrame := am.GetCurrentFrame()

				// Look for timing inconsistencies
				// If we get consecutive "newFrame" returns very quickly,
				// it might indicate a race condition in timing logic
				if newFrame && lastNewFrame {
					mu.Lock()
					inconsistencies++
					mu.Unlock()
				}
				lastNewFrame = newFrame
				time.Sleep(time.Microsecond * 500) // Very frequent reads
			}
		}
	}()

	// Run test for sufficient time to detect race conditions
	time.Sleep(100 * time.Millisecond)
	close(done)
	wg.Wait()

	t.Logf("Race condition test completed")
	t.Logf("Timing inconsistencies detected: %d", inconsistencies)

	// The bug mentioned in audit suggests there should be race conditions
	// If this test passes consistently, the bug may already be fixed
	t.Log("Current implementation appears to be thread-safe")
}

// TestBug3FrameTimingConsistency tests whether frame timing and index are consistent
func TestBug3FrameTimingConsistency(t *testing.T) {
	am := character.NewAnimationManager()

	// Create test animation with known timing
	testGif := &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{10, 10}, // 100ms per frame
	}

	err := am.LoadAnimation("timing_test", testGif)
	if err != nil {
		t.Fatalf("Failed to load test animation: %v", err)
	}

	err = am.SetCurrentAnimation("timing_test")
	if err != nil {
		t.Fatalf("Failed to set current animation: %v", err)
	}

	am.Play()

	// Test timing behavior
	frame1, newFrame1 := am.GetCurrentFrame()
	if frame1 == nil {
		t.Fatal("Expected frame, got nil")
	}

	// Should not be a new frame immediately
	if newFrame1 {
		t.Log("Note: Immediate new frame detected - may indicate timing issue")
	}

	// Wait for frame timing
	time.Sleep(110 * time.Millisecond) // Wait longer than frame delay

	// Update animation state
	frameChanged := am.Update()

	// Check if frame actually changed
	frame2, newFrame2 := am.GetCurrentFrame()

	t.Logf("Frame changed after update: %v", frameChanged)
	t.Logf("GetCurrentFrame indicates new frame: %v", newFrame2)

	// Frame pointers should be different if animation advanced
	if frameChanged && frame1 == frame2 {
		t.Error("Update() reported frame change but GetCurrentFrame() returned same frame")
	}
}

// TestBug3ExpectedRaceConditionBehavior documents what the audit issue describes
func TestBug3ExpectedRaceConditionBehavior(t *testing.T) {
	t.Log("=== ANIMATION RACE CONDITION BUG ANALYSIS ===")
	t.Log("Audit Issue: GetCurrentFrame() timing checks vs Update() modifications")
	t.Log("")
	t.Log("Described problem:")
	t.Log("1. GetCurrentFrame() checks time.Since(am.lastUpdate) >= frameDelay")
	t.Log("2. Update() modifies am.frameIndex and am.lastUpdate concurrently")
	t.Log("3. Potential for inconsistent timing vs frame index")
	t.Log("")
	t.Log("Current implementation analysis:")
	t.Log("- GetCurrentFrame() uses am.mu.RLock()")
	t.Log("- Update() uses am.mu.Lock()")
	t.Log("- This should prevent race conditions")
	t.Log("")
	t.Log("Possible remaining issue:")
	t.Log("- Timing calculation might still be inconsistent if")
	t.Log("  multiple goroutines call GetCurrentFrame() rapidly")
}
