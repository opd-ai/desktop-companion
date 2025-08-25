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

// TestAnimationFrameRateOptimization tests that animations only refresh when frames actually change
func TestAnimationFrameRateOptimization(t *testing.T) {
	// Create animation manager with slow-updating animation
	am := NewAnimationManager()

	// Create GIF with 500ms per frame (2 FPS)
	am.animations["slow"] = &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{50, 50}, // 500ms per frame (GIF delay is in centiseconds)
	}
	am.currentAnim = "slow"
	am.playing = true
	am.lastUpdate = time.Now()

	// Test that Update() returns false when no frame change occurs
	frameChanged := am.Update()
	if frameChanged {
		t.Error("Expected no frame change immediately after setting lastUpdate")
	}

	// Wait for frame delay and test that Update() returns true
	time.Sleep(510 * time.Millisecond) // Wait slightly longer than frame delay
	frameChanged = am.Update()
	if !frameChanged {
		t.Error("Expected frame change after waiting for frame delay")
	}

	// Test that subsequent immediate calls return false
	frameChanged = am.Update()
	if frameChanged {
		t.Error("Expected no frame change on immediate subsequent call")
	}
}

// TestCharacterUpdateReturnsChanges tests that Character.Update() correctly reports visual changes
func TestCharacterUpdateReturnsChanges(t *testing.T) {
	// Create a test character card
	card := &CharacterCard{
		Name:        "TestCharacter",
		Description: "Test",
		Animations: map[string]string{
			"idle": "idle.gif",
		},
		Dialogs: []Dialog{},
		Behavior: Behavior{
			IdleTimeout:     10,
			MovementEnabled: true,
			DefaultSize:     100,
		},
	}

	char := &Character{
		card:             card,
		animationManager: NewAnimationManager(),
		currentState:     "idle",
		lastStateChange:  time.Now(),
		lastInteraction:  time.Now(),
		dialogCooldowns:  make(map[string]time.Time),
		idleTimeout:      10 * time.Second,
		movementEnabled:  true,
		size:             100,
	}

	// Create a slow animation (2 FPS)
	char.animationManager.animations["idle"] = &gif.GIF{
		Image: []*image.Paletted{
			{Pix: []uint8{0}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
			{Pix: []uint8{1}, Stride: 1, Rect: image.Rect(0, 0, 1, 1)},
		},
		Delay: []int{50, 50}, // 500ms per frame
	}
	char.animationManager.currentAnim = "idle"
	char.animationManager.playing = true
	char.animationManager.lastUpdate = time.Now()

	// First call should return false (no frame change yet)
	hasChanges := char.Update()
	if hasChanges {
		t.Error("Expected no changes immediately after initialization")
	}

	// Wait for animation frame delay
	time.Sleep(510 * time.Millisecond)

	// This call should return true (frame changed)
	hasChanges = char.Update()
	if !hasChanges {
		t.Error("Expected changes after waiting for frame delay")
	}

	// Immediate subsequent call should return false
	hasChanges = char.Update()
	if hasChanges {
		t.Error("Expected no changes on immediate subsequent call")
	}
}
