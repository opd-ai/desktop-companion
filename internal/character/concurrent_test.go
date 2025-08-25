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

// TestDialogCooldownRaceCondition reproduces the race condition in HandleHover
func TestDialogCooldownRaceCondition(t *testing.T) {
	// Create a character with hover dialog
	card := &CharacterCard{
		Name:        "TestCharacter",
		Description: "Test",
		Animations: map[string]string{
			"idle": "idle.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "hover",
				Animation: "idle",
				Responses: []string{"Hello!"},
				Cooldown:  1, // 1 second cooldown
			},
			{
				Trigger:   "click",
				Animation: "idle",
				Responses: []string{"Clicked!"},
				Cooldown:  1,
			},
		},
		Behavior: Behavior{
			IdleTimeout:     10,
			MovementEnabled: true,
			DefaultSize:     100,
		},
	}

	char := &Character{
		card:            card,
		dialogCooldowns: make(map[string]time.Time),
		currentState:    "idle",
		lastStateChange: time.Now(),
		lastInteraction: time.Now().Add(-5 * time.Second), // Make sure hover can trigger
		movementEnabled: true,
		size:            100,
	}

	var wg sync.WaitGroup
	done := make(chan struct{})
	responses := make(chan string, 100)

	// Start multiple hover calls (simulates rapid mouse movement)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					response := char.HandleHover() // This should now use proper synchronization
					if response != "" {
						responses <- response
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	// Start click calls that modify cooldowns (simulates user interactions)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				char.HandleClick() // Uses write lock and updates cooldowns
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	// Run for a short time to allow race conditions to manifest
	time.Sleep(100 * time.Millisecond)
	close(done)

	wg.Wait()
	close(responses)

	// Count responses - with proper fix, hover responses should be limited by cooldowns
	responseCount := 0
	for range responses {
		responseCount++
	}

	// With proper synchronization, hover responses should be limited by cooldowns
	if responseCount > 20 { // Arbitrary high threshold indicating potential race
		t.Logf("Warning: High response count (%d) may indicate cooldown race condition", responseCount)
	}
}
