package character

import (
	"image"
	"image/gif"
	"sync"
	"testing"
	"time"
)

// TestConcurrentFrameUpdates reproduces the race condition in GetCurrentFrame
func TestConcurrentFrameUpdates(t *testing.T) {
	// Create animation manager with test data
	am := NewAnimationManager()

	// Create a minimal GIF structure for testing
	// This simulates loaded animation data
	am.animations["test"] = &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{10, 10}, // 100ms per frame
	}
	am.currentAnim = "test"
	am.playing = true

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Start Update loop (simulates animation loop)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				am.Update() // This uses write lock
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	// Start GetCurrentFrame calls (simulates renderer)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				am.GetCurrentFrame() // This should use read lock but modifies state
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// Run for a short time to allow race conditions to manifest
	time.Sleep(50 * time.Millisecond)
	close(done)

	wg.Wait()
}
