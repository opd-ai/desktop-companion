package character

import (
	"testing"
	"time"
)

// TestAdvancedFeaturesBasicFunctionality tests basic integration of Phase 3 Task 3 features
func TestAdvancedFeaturesBasicFunctionality(t *testing.T) {
	// Create a complete romance character
	card := createRomanceCharacterCard()
	char := createTestCharacterWithRomanceFeatures(card, true)

	gameState := char.GetGameState()
	if gameState == nil {
		t.Fatal("Game state should not be nil")
	}

	t.Run("advanced_features_initialized", func(t *testing.T) {
		// Test that all advanced features are initialized
		if char.jealousyManager == nil {
			t.Error("Jealousy manager should be initialized")
		}

		if char.compatibilityAnalyzer == nil {
			t.Error("Compatibility analyzer should be initialized")
		}

		if char.crisisRecoveryManager == nil {
			t.Error("Crisis recovery manager should be initialized")
		}
	})

	t.Run("jealousy_manager_basic_operation", func(t *testing.T) {
		if char.jealousyManager == nil {
			t.Skip("Jealousy manager not initialized")
		}

		// Test basic jealousy manager functions
		level := char.jealousyManager.GetJealousyLevel(gameState)
		if level < 0 || level > 1 {
			t.Errorf("Jealousy level should be 0-1, got %f", level)
		}

		status := char.jealousyManager.GetStatus(gameState)
		if status == nil {
			t.Error("Status should not be nil")
		}

		if enabled, ok := status["enabled"]; !ok || !enabled.(bool) {
			t.Error("Jealousy manager should be enabled for romance characters")
		}
	})

	t.Run("compatibility_analyzer_basic_operation", func(t *testing.T) {
		if char.compatibilityAnalyzer == nil {
			t.Skip("Compatibility analyzer not initialized")
		}

		// Test basic compatibility analyzer functions
		insights := char.compatibilityAnalyzer.GetCompatibilityInsights()
		if insights == nil {
			t.Error("Insights should not be nil")
		}

		if enabled, ok := insights["enabled"]; !ok || !enabled.(bool) {
			t.Error("Compatibility analyzer should be enabled")
		}

		pattern := char.compatibilityAnalyzer.GetPlayerPattern()
		if pattern == nil {
			t.Log("No player pattern available yet (expected for new character)")
		}
	})

	t.Run("crisis_recovery_manager_basic_operation", func(t *testing.T) {
		if char.crisisRecoveryManager == nil {
			t.Skip("Crisis recovery manager not initialized")
		}

		// Test basic crisis recovery manager functions
		status := char.crisisRecoveryManager.GetCrisisStatus()
		if status == nil {
			t.Error("Crisis status should not be nil")
		}

		if enabled, ok := status["enabled"]; !ok || !enabled.(bool) {
			t.Error("Crisis recovery manager should be enabled")
		}

		activeCrises := char.crisisRecoveryManager.GetActiveCrises()
		if activeCrises == nil {
			t.Error("Active crises list should not be nil (can be empty)")
		}
	})

	t.Run("character_update_processes_advanced_features", func(t *testing.T) {
		// Set up some conditions that might trigger advanced features
		gameState.ApplyInteractionEffects(map[string]float64{
			"affection": 30.0,
			"jealousy":  50.0,
		})

		// Simulate some time passage
		char.lastInteraction = time.Now().Add(-30 * time.Minute)

		// Run character update multiple times
		for i := 0; i < 5; i++ {
			changed := char.Update()
			// changed can be true or false, both are valid
			_ = changed
			time.Sleep(10 * time.Millisecond)
		}

		// Verify the character is still functional
		finalState := char.GetGameState()
		if finalState == nil {
			t.Error("Game state should remain available after updates")
		}

		// Check that stats are within valid ranges
		for statName, stat := range finalState.Stats {
			if stat.Current < 0 || stat.Current > stat.Max {
				t.Errorf("Stat %s out of range: %f (max: %f)", statName, stat.Current, stat.Max)
			}
		}
	})

	t.Run("romance_interactions_work_with_advanced_features", func(t *testing.T) {
		// Test that romance interactions still work with advanced features enabled
		initialAffection := gameState.GetStat("affection")

		response := char.HandleRomanceInteraction("compliment")
		if response == "" {
			t.Error("Romance interaction should return a response")
		}

		finalAffection := gameState.GetStat("affection")
		if finalAffection <= initialAffection {
			t.Error("Compliment should increase affection")
		}

		// Test that compatibility analysis might kick in
		if char.compatibilityAnalyzer != nil {
			// Multiple interactions to build pattern
			for i := 0; i < 5; i++ {
				char.HandleRomanceInteraction("compliment")
				time.Sleep(10 * time.Millisecond)
			}

			pattern := char.compatibilityAnalyzer.GetPlayerPattern()
			if pattern != nil && pattern.TotalInteractions > 0 {
				t.Logf("Compatibility analysis working: %d interactions tracked", pattern.TotalInteractions)
			}
		}
	})

	t.Run("crisis_scenario_basic_test", func(t *testing.T) {
		if char.crisisRecoveryManager == nil {
			t.Skip("Crisis recovery manager not initialized")
		}

		// Create a high jealousy scenario
		gameState.ApplyInteractionEffects(map[string]float64{
			"jealousy": 85.0, // High jealousy
		})

		// Update character to process the crisis
		char.Update()

		// Check if crisis is detected
		activeCrises := char.crisisRecoveryManager.GetActiveCrises()
		if len(activeCrises) > 0 {
			t.Logf("Crisis detected: %s", activeCrises[0].Name)

			// Test crisis recovery interaction
			if char.card.Interactions != nil {
				if _, hasApology := char.card.Interactions["apology"]; hasApology {
					response := char.HandleRomanceInteraction("apology")
					if response == "" {
						t.Error("Apology interaction should work during crisis")
					}
				}
			}
		} else {
			t.Log("No crisis detected - this can be normal depending on thresholds")
		}
	})
}

// TestAdvancedFeaturesConfiguration tests that the advanced features are properly configured
func TestAdvancedFeaturesConfiguration(t *testing.T) {
	card := createRomanceCharacterCard()

	// Verify the card has required configuration for advanced features
	if card.Personality == nil {
		t.Error("Romance character should have personality configuration")
	}

	if card.Personality != nil {
		requiredTraits := []string{"jealousy_prone", "affection_responsiveness", "trust_difficulty"}
		for _, trait := range requiredTraits {
			if _, exists := card.Personality.Traits[trait]; !exists {
				t.Errorf("Personality should include trait: %s", trait)
			}
		}
	}

	// Verify crisis recovery interactions are available
	if card.Interactions == nil {
		t.Error("Character should have interactions configured")
	} else {
		recoveryInteractions := []string{"apology", "reassurance"}
		for _, interaction := range recoveryInteractions {
			if _, exists := card.Interactions[interaction]; !exists {
				t.Errorf("Character should have crisis recovery interaction: %s", interaction)
			}
		}
	}

	// Test character creation with the configuration
	char := createTestCharacterWithRomanceFeatures(card, true)
	if char == nil {
		t.Fatal("Character creation should succeed with proper configuration")
	}

	// Verify advanced features are properly initialized
	if char.jealousyManager == nil {
		t.Error("Jealousy manager should be initialized for configured romance character")
	}

	if char.compatibilityAnalyzer == nil {
		t.Error("Compatibility analyzer should be initialized for configured romance character")
	}

	if char.crisisRecoveryManager == nil {
		t.Error("Crisis recovery manager should be initialized for configured romance character")
	}
}
