package character

import (
	"testing"
	"time"
)

// TestEventFrequencyMultiplier_GetterSetter tests basic getter and setter functionality
func TestEventFrequencyMultiplier_GetterSetter(t *testing.T) {
	// Create test character
	char, err := createTestCharacter()
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Test default value
	defaultMultiplier := char.GetEventFrequencyMultiplier()
	if defaultMultiplier != 1.0 {
		t.Errorf("Expected default frequency multiplier to be 1.0, got %.1f", defaultMultiplier)
	}

	// Test setting valid values
	testValues := []float64{0.5, 1.0, 1.5, 2.0, 3.0}
	for _, value := range testValues {
		char.SetEventFrequencyMultiplier(value)
		result := char.GetEventFrequencyMultiplier()
		if result != value {
			t.Errorf("Expected frequency multiplier %.1f, got %.1f", value, result)
		}
	}
}

// TestEventFrequencyMultiplier_Clamping tests value clamping to valid range
func TestEventFrequencyMultiplier_Clamping(t *testing.T) {
	char, err := createTestCharacter()
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Test values below minimum
	char.SetEventFrequencyMultiplier(0.05)
	result := char.GetEventFrequencyMultiplier()
	if result != 0.1 {
		t.Errorf("Expected clamped value 0.1, got %.1f", result)
	}

	char.SetEventFrequencyMultiplier(-1.0)
	result = char.GetEventFrequencyMultiplier()
	if result != 0.1 {
		t.Errorf("Expected clamped value 0.1, got %.1f", result)
	}

	// Test values above maximum
	char.SetEventFrequencyMultiplier(4.0)
	result = char.GetEventFrequencyMultiplier()
	if result != 3.0 {
		t.Errorf("Expected clamped value 3.0, got %.1f", result)
	}

	char.SetEventFrequencyMultiplier(10.0)
	result = char.GetEventFrequencyMultiplier()
	if result != 3.0 {
		t.Errorf("Expected clamped value 3.0, got %.1f", result)
	}
}

// TestEventFrequencyMultiplier_ThreadSafety tests concurrent access
func TestEventFrequencyMultiplier_ThreadSafety(t *testing.T) {
	char, err := createTestCharacter()
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	// Goroutine 1: Writing values
	go func() {
		for i := 0; i < 100; i++ {
			char.SetEventFrequencyMultiplier(1.5)
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Goroutine 2: Reading values
	go func() {
		for i := 0; i < 100; i++ {
			_ = char.GetEventFrequencyMultiplier()
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done

	// Verify final state
	result := char.GetEventFrequencyMultiplier()
	if result != 1.5 {
		t.Errorf("Expected final frequency multiplier 1.5, got %.1f", result)
	}
}

// TestHasRandomEvents tests random events detection
func TestHasRandomEvents(t *testing.T) {
	// Test character without random events
	char, err := createTestCharacter()
	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	if char.HasRandomEvents() {
		t.Error("Expected character without random events to return false from HasRandomEvents()")
	}

	// Test character with random events - need to set up stats and random events
	char.card.Stats = map[string]StatConfig{
		"happiness": {Initial: 100, Max: 100, DegradationRate: 0.8, CriticalThreshold: 15},
	}
	char.card.RandomEvents = []RandomEventConfig{
		{
			Name:        "test_event",
			Description: "Test event",
			Probability: 0.1,
			Effects: map[string]float64{
				"happiness": 5.0,
			},
		},
	}

	// Initialize game features to set up random event manager
	char.initializeGameFeatures()

	if !char.HasRandomEvents() {
		t.Error("Expected character with random events to return true from HasRandomEvents()")
	}
}

// TestRandomEventManager_UpdateWithFrequency tests the new frequency-aware update method
func TestRandomEventManager_UpdateWithFrequency(t *testing.T) {
	events := []RandomEventConfig{
		{
			Name:        "test_event",
			Description: "Test event for frequency testing",
			Probability: 0.1, // 10% base probability
			Effects: map[string]float64{
				"happiness": 5.0,
			},
			Cooldown: 1, // 1 second cooldown
		},
	}

	rem := NewRandomEventManager(events, true, time.Millisecond*100)
	gameState := createTestGameState()

	// Test with different frequency multipliers
	testCases := []struct {
		multiplier float64
		name       string
	}{
		{0.5, "half frequency"},
		{1.0, "normal frequency"},
		{1.5, "1.5x frequency"},
		{2.0, "double frequency"},
		{3.0, "triple frequency"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset event manager state
			rem.eventCooldowns = make(map[string]time.Time)
			rem.lastCheck = time.Now().Add(-time.Hour) // Force check

			// Call UpdateWithFrequency
			result := rem.UpdateWithFrequency(time.Second, gameState, tc.multiplier)

			// We can't guarantee an event will trigger due to randomness,
			// but we can verify the method doesn't crash and returns expected type
			if result != nil && result.Name != "test_event" {
				t.Errorf("Expected event name 'test_event', got '%s'", result.Name)
			}
		})
	}
}

// TestRandomEventManager_FrequencyProbabilityCalculation tests probability adjustment
func TestRandomEventManager_FrequencyProbabilityCalculation(t *testing.T) {
	events := []RandomEventConfig{
		{
			Name:        "test_event",
			Description: "Test event",
			Probability: 0.8, // 80% base probability for reliable testing
			Effects: map[string]float64{
				"happiness": 5.0,
			},
			Cooldown: 0, // No cooldown for testing
		},
	}

	rem := NewRandomEventManager(events, true, time.Millisecond*10)
	gameState := createTestGameState()

	// Test probability capping at 1.0
	rem.lastCheck = time.Now().Add(-time.Hour) // Force check

	// With 80% base probability and 2x multiplier, should cap at 100%
	// We can't test exact probability due to randomness, but we can verify
	// the method handles the calculation correctly
	result := rem.UpdateWithFrequency(time.Second, gameState, 2.0)

	// Just verify the method works - actual probability testing would require
	// many iterations and statistical analysis
	_ = result // Method should not panic
}

// TestRandomEventManager_BackwardCompatibility tests that original Update method still works
func TestRandomEventManager_BackwardCompatibility(t *testing.T) {
	events := []RandomEventConfig{
		{
			Name:        "test_event",
			Description: "Test event",
			Probability: 0.1,
			Effects: map[string]float64{
				"happiness": 5.0,
			},
		},
	}

	rem := NewRandomEventManager(events, true, time.Millisecond*100)
	gameState := createTestGameState()

	// Original Update method should still work
	rem.lastCheck = time.Now().Add(-time.Hour) // Force check
	result := rem.Update(time.Second, gameState)

	// Verify method doesn't crash
	_ = result
}

// Helper function to create a test character
func createTestCharacter() (*Character, error) {
	card := &CharacterCard{
		Name:        "Test Character",
		Description: "A character for testing",
		Animations: map[string]string{
			"idle": "test.gif",
		},
		Behavior: Behavior{
			IdleTimeout: 30,
			DefaultSize: 128,
		},
	}

	return New(card, "")
}
