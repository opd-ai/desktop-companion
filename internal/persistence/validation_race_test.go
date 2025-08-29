package persistence

import (
	"sync"
	"testing"
	"time"
)

// TestSaveDataValidationRaceCondition attempts to reproduce the race condition
// described in AUDIT.md where validateSaveData accesses data fields without
// synchronization while concurrent auto-save operations may be modifying the same data
func TestSaveDataValidationRaceCondition(t *testing.T) {
	sm := NewSaveManager(t.TempDir())
	defer sm.Close()

	// Create initial save data
	now := time.Now()
	testData := &GameSaveData{
		CharacterName: "RaceTestCharacter",
		SaveVersion:   "1.0",
		GameState: &GameStateData{
			Stats: map[string]*StatData{
				"hunger": {
					Current:           75.0,
					Max:               100.0,
					DegradationRate:   0.1,
					CriticalThreshold: 20.0,
				},
			},
			CreationTime:    now,
			LastDecayUpdate: now,
		},
	}

	// Save initial data
	if err := sm.SaveGameState("RaceTestCharacter", testData); err != nil {
		t.Fatalf("Failed to save initial data: %v", err)
	}

	// Enable auto-save with a very short interval to increase chance of race conditions
	gameStateProvider := func() *GameSaveData {
		// Modify data each time it's accessed (simulating ongoing game state changes)
		now := time.Now()
		return &GameSaveData{
			CharacterName: "RaceTestCharacter",
			SaveVersion:   "1.0",
			GameState: &GameStateData{
				Stats: map[string]*StatData{
					"hunger": {
						Current:           float64(time.Now().UnixNano() % 100), // Constantly changing
						Max:               100.0,
						DegradationRate:   0.1,
						CriticalThreshold: 20.0,
					},
				},
				CreationTime:    now,
				LastDecayUpdate: now,
			},
		}
	}

	sm.EnableAutoSave(10*time.Millisecond, gameStateProvider)
	defer sm.DisableAutoSave()

	// Run concurrent load operations while auto-save is running
	// This should trigger the race condition if it exists
	var wg sync.WaitGroup
	errors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				// Load the game state - this calls validateSaveData
				data, err := sm.LoadGameState("RaceTestCharacter")
				if err != nil {
					errors <- err
					return
				}

				if data == nil {
					errors <- err
					return
				}

				// Brief pause to allow auto-save to continue
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// Wait for all load operations to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		t.Log("All load operations completed successfully")
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - possible deadlock or race condition")
	}

	// Check for any errors
	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("Load operation failed: %v", err)
		}
	}
}

// TestValidationWithSharedData tests if validation could have race conditions
// when multiple goroutines validate the same data structure
func TestValidationWithSharedData(t *testing.T) {
	sm := NewSaveManager(t.TempDir())

	// Create shared data that will be validated concurrently
	now := time.Now()
	sharedData := &GameSaveData{
		CharacterName: "SharedTestCharacter",
		SaveVersion:   "1.0",
		GameState: &GameStateData{
			Stats: map[string]*StatData{
				"hunger": {
					Current:           50.0,
					Max:               100.0,
					DegradationRate:   0.1,
					CriticalThreshold: 20.0,
				},
			},
			CreationTime:    now,
			LastDecayUpdate: now,
		},
	}

	// Run concurrent validation while modifying the data
	var wg sync.WaitGroup
	errors := make(chan error, 20)

	// Goroutine that continuously modifies the shared data
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			// Modify the data structure while validation might be reading it
			sharedData.GameState.Stats["hunger"].Current = float64(i)
			sharedData.CharacterName = "Modified" + string(rune(i))
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Multiple goroutines that validate the same data structure
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				// This should demonstrate the race condition if it exists
				if err := sm.validateSaveData(sharedData); err != nil {
					// Validation errors are expected due to the modifications
					// We're looking for panics or data races, not validation failures
				}
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	close(errors)

	// If we reach here without panics, the race condition might not exist
	// or might be very rare
	t.Log("Concurrent validation test completed without crashes")
}
