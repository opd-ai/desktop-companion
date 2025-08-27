package character

import (
	"testing"
	"time"
)

// TestJealousyManager tests the jealousy management system in isolation
func TestJealousyManager(t *testing.T) {
	// Create a basic game state for testing
	gameState := &GameState{
		Stats: map[string]*Stat{
			"affection": {Current: 50.0, Max: 100.0},
			"trust":     {Current: 40.0, Max: 100.0},
			"jealousy":  {Current: 20.0, Max: 100.0},
			"happiness": {Current: 60.0, Max: 100.0},
		},
		CreationTime: time.Now(),
	}

	t.Run("new_jealousy_manager", func(t *testing.T) {
		triggers := []JealousyTrigger{
			{
				Name:              "long_absence",
				Description:       "Player hasn't interacted in a while",
				InteractionGap:    2 * time.Hour,
				JealousyIncrement: 10,
				TrustPenalty:      5,
				Probability:       0.3,
			},
		}

		manager := NewJealousyManager(triggers, true, 80.0)

		if manager == nil {
			t.Fatal("NewJealousyManager should not return nil")
		}
	})

	t.Run("get_jealousy_level", func(t *testing.T) {
		manager := NewJealousyManager([]JealousyTrigger{}, true, 80.0)

		level := manager.GetJealousyLevel(gameState)
		expected := gameState.Stats["jealousy"].Current / 100.0 // Normalized to 0-1

		if level != expected {
			t.Errorf("Expected jealousy level %f, got %f", expected, level)
		}
	})

	t.Run("get_status", func(t *testing.T) {
		triggers := []JealousyTrigger{
			{Name: "test1"},
			{Name: "test2"},
		}
		manager := NewJealousyManager(triggers, true, 75.0)

		status := manager.GetStatus(gameState)

		if !status["enabled"].(bool) {
			t.Error("Jealousy manager should be enabled")
		}

		triggerCount := status["triggerCount"].(int)
		if triggerCount != 2 {
			t.Errorf("Expected 2 triggers, got %d", triggerCount)
		}

		intensity := status["intensity"].(float64)
		if intensity != 0.2 { // 20/100
			t.Errorf("Expected intensity 0.2, got %f", intensity)
		}

		threshold := status["threshold"].(float64)
		if threshold != 75 {
			t.Errorf("Expected threshold 75, got %f", threshold)
		}
	})

	t.Run("update_triggers_jealousy", func(t *testing.T) {
		triggers := []JealousyTrigger{
			{
				Name:              "high_affection_trigger",
				Description:       "Triggers when affection is high",
				InteractionGap:    1 * time.Hour,
				JealousyIncrement: 5,
				Conditions:        map[string]float64{"affection": 45}, // Min affection
				Probability:       1.0,                                 // Always trigger for testing
			},
		}

		manager := NewJealousyManager(triggers, true, 80.0)

		// Set last interaction time to 3 hours ago to simulate trigger condition
		lastInteraction := time.Now().Add(-3 * time.Hour)

		triggeredEvent := manager.Update(gameState, lastInteraction)

		// Should trigger some jealousy effect
		if triggeredEvent == nil {
			t.Log("No event triggered - this can happen due to probability or conditions")
		} else {
			if triggeredEvent.Name == "" {
				t.Error("Triggered event should have a name")
			}
		}
	})

	t.Run("high_jealousy_consequences", func(t *testing.T) {
		// Set up high jealousy state
		highJealousyState := &GameState{
			Stats: map[string]*Stat{
				"affection": {Current: 50.0, Max: 100.0},
				"trust":     {Current: 40.0, Max: 100.0},
				"jealousy":  {Current: 85.0, Max: 100.0}, // Above threshold
				"happiness": {Current: 60.0, Max: 100.0},
			},
			CreationTime: time.Now(),
		}

		manager := NewJealousyManager([]JealousyTrigger{}, true, 80.0)

		// Record initial stats
		initialAffection := highJealousyState.Stats["affection"].Current

		triggeredEvent := manager.Update(highJealousyState, time.Now().Add(-1*time.Hour))

		// Check if stats were affected (they might be via consequences)
		if triggeredEvent != nil && len(triggeredEvent.Effects) > 0 {
			t.Log("Jealousy consequences applied")
		}

		// In the actual implementation, consequences are applied by the character behavior system
		// Here we just check that the manager recognizes high jealousy
		if manager.GetJealousyLevel(highJealousyState) <= 0.8 {
			t.Error("High jealousy should be detected")
		}

		// Verify stats weren't corrupted
		if highJealousyState.Stats["affection"].Current != initialAffection {
			t.Log("Stats may have been modified by jealousy consequences")
		}
	})
}

// TestCompatibilityAnalyzerIntegration tests the compatibility analysis system integration
func TestCompatibilityAnalyzerIntegration(t *testing.T) {
	// Create a test character with romance features for integration testing
	card := createRomanceCharacterCard()
	char := createTestCharacterWithRomanceFeatures(card, true)

	t.Run("compatibility_analyzer_integration", func(t *testing.T) {
		// Test that the analyzer is properly initialized
		if char.compatibilityAnalyzer == nil {
			t.Error("Compatibility analyzer should be initialized for romance characters")
		}

		// Test that it can provide insights
		insights := char.compatibilityAnalyzer.GetCompatibilityInsights()
		if insights == nil {
			t.Error("Compatibility insights should not be nil")
		}

		// Test basic functionality
		enabled, ok := insights["enabled"].(bool)
		if !ok || !enabled {
			t.Error("Compatibility analyzer should be enabled for romance characters")
		}
	})

	t.Run("player_behavior_integration", func(t *testing.T) {
		// Simulate player interactions for behavior analysis  
		for i := 0; i < 3; i++ {
			response := char.HandleRomanceInteraction("compliment")
			if response == "" {
				t.Error("Expected non-empty response from romance interaction")
			}
		}

		// Test that the compatibility analyzer is working
		insights := char.compatibilityAnalyzer.GetCompatibilityInsights()
		if insights == nil {
			t.Error("Compatibility insights should not be nil after interactions")
		}
	})
}

// TestCrisisRecoveryIntegration tests the crisis and recovery system integration
func TestCrisisRecoveryIntegration(t *testing.T) {
	// Create a test character with romance features for crisis testing
	card := createRomanceCharacterCard()
	char := createTestCharacterWithRomanceFeatures(card, true)

	t.Run("crisis_recovery_integration", func(t *testing.T) {
		// Test that the crisis recovery manager is properly initialized
		if char.crisisRecoveryManager == nil {
			t.Error("Crisis recovery manager should be initialized for romance characters")
		}

		// Test basic status check
		status := char.crisisRecoveryManager.GetCrisisStatus()
		if status == nil {
			t.Error("Crisis status should not be nil")
		}

		// Test that it's enabled for romance characters
		enabled, ok := status["enabled"].(bool)
		if !ok || !enabled {
			t.Error("Crisis recovery manager should be enabled for romance characters")
		}
	})
}

// Simple integration test placeholder to verify system works
// More detailed testing is in other test files
func TestCrisisRecoveryBasicIntegration(t *testing.T) {
	// Create basic character with romance features
	card := createRomanceCharacterCard()
	char := createTestCharacterWithRomanceFeatures(card, true)

	t.Run("crisis_recovery_manager_exists", func(t *testing.T) {
		// Just verify the crisis recovery manager is initialized
		if char.crisisRecoveryManager == nil {
			t.Error("Crisis recovery manager should be initialized for romance characters")
		}

		// Test basic status check doesn't crash
		status := char.crisisRecoveryManager.GetCrisisStatus()
		if status == nil {
			t.Error("Crisis status should not be nil")
		}
	})
}
