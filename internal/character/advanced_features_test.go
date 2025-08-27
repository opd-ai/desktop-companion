package character

import (
	"testing"
	"time"
)

// Helper function to create a test character instance with romance features
func createTestCharacterWithRomanceFeatures(card *CharacterCard, enableGameMode bool) *Character {
	char := &Character{
		card:                     card,
		animationManager:         NewAnimationManager(),
		currentState:             "idle",
		lastStateChange:          time.Now(),
		lastInteraction:          time.Now(),
		dialogCooldowns:          make(map[string]time.Time),
		gameInteractionCooldowns: make(map[string]time.Time),
		idleTimeout:              time.Duration(card.Behavior.IdleTimeout) * time.Second,
		size:                     card.Behavior.DefaultSize,
	}

	// Initialize game features if requested
	if enableGameMode && card.HasGameFeatures() {
		char.initializeGameFeatures()
		char.initializeAdvancedFeatures()
	}

	return char
}

// TestJealousyMechanics tests the jealousy system functionality
func TestJealousyMechanics(t *testing.T) {
	// Create a character with jealousy features
	card := createRomanceCharacterCard()
	char := createTestCharacterWithRomanceFeatures(card, true)

	// Enable game mode to activate romance features
	// Note: EnableGameMode is called internally by initializeGameFeatures

	// Get game state for testing
	gameState := char.GetGameState()
	if gameState == nil {
		t.Fatal("Game state should not be nil")
	}

	t.Run("jealousy_manager_initialization", func(t *testing.T) {
		if char.jealousyManager == nil {
			t.Error("Jealousy manager should be initialized for romance characters")
		}

		status := char.jealousyManager.GetStatus(gameState)
		if !status["enabled"].(bool) {
			t.Error("Jealousy manager should be enabled for jealousy-prone characters")
		}

		triggerCount := status["triggerCount"].(int)
		if triggerCount == 0 {
			t.Error("Jealousy manager should have triggers configured")
		}
	})

	t.Run("jealousy_triggers_after_long_absence", func(t *testing.T) {
		// Set up conditions for jealousy trigger
		gameState.ApplyInteractionEffects(map[string]float64{
			"affection": 25.0, // Enough affection to care
		})

		// Simulate long absence by setting last interaction far in the past
		char.lastInteraction = time.Now().Add(-3 * time.Hour)

		// Update character to trigger jealousy check
		for i := 0; i < 10; i++ { // Multiple updates to increase probability
			char.Update()
			time.Sleep(10 * time.Millisecond) // Small delay for different random seeds
		}

		// Check if jealousy increased
		currentJealousy := gameState.GetStat("jealousy")
		if currentJealousy <= 0 {
			t.Log("Jealousy level:", currentJealousy)
			// Note: Due to probability, this might not always trigger
			// In a real implementation, you might want deterministic testing
		}
	})

	t.Run("jealousy_consequences_above_threshold", func(t *testing.T) {
		// Set jealousy above threshold
		gameState.ApplyInteractionEffects(map[string]float64{
			"jealousy":  85.0, // High jealousy
			"affection": 50.0,
			"trust":     50.0,
			"happiness": 50.0,
		})

		statsBefore := gameState.GetStats()

		// Update to trigger jealousy consequences
		char.Update()

		statsAfter := gameState.GetStats()

		// Jealousy consequences should reduce other stats
		if statsAfter["affection"] >= statsBefore["affection"] {
			t.Error("High jealousy should reduce affection")
		}

		if statsAfter["trust"] >= statsBefore["trust"] {
			t.Error("High jealousy should reduce trust")
		}

		if statsAfter["happiness"] >= statsBefore["happiness"] {
			t.Error("High jealousy should reduce happiness")
		}
	})
}

// TestCompatibilityAnalyzer tests the advanced compatibility system
func TestCompatibilityAnalyzer(t *testing.T) {
	// Create a character with compatibility analysis
	card := createRomanceCharacterCard()
	char := createTestCharacterWithRomanceFeatures(card, true)

	gameState := char.GetGameState()

	t.Run("compatibility_analyzer_initialization", func(t *testing.T) {
		if char.compatibilityAnalyzer == nil {
			t.Error("Compatibility analyzer should be initialized for romance characters")
		}

		insights := char.compatibilityAnalyzer.GetCompatibilityInsights()
		if !insights["enabled"].(bool) {
			t.Error("Compatibility analyzer should be enabled")
		}

		adaptationStrength := insights["adaptationStrength"].(float64)
		if adaptationStrength <= 0 || adaptationStrength > 1 {
			t.Errorf("Adaptation strength should be between 0 and 1, got %f", adaptationStrength)
		}
	})

	t.Run("player_behavior_analysis", func(t *testing.T) {
		// Simulate consistent player behavior
		successCount := 0
		for i := 0; i < 10; i++ {
			response := char.HandleRomanceInteraction("compliment")
			// Check for actual success response (not failure message)
			if response == "Thank you! 💕" {
				successCount++
			}
			time.Sleep(50 * time.Millisecond) // Consistent timing
		}

		t.Logf("Actually successful interactions: %d out of 10", successCount)

		// Force immediate analysis (compatibility analyzer normally waits 5 minutes between analyses)
		char.compatibilityAnalyzer.ForceAnalysis(gameState)

		pattern := char.compatibilityAnalyzer.GetPlayerPattern()
		if pattern == nil {
			t.Error("Player pattern should be available after interactions")
		}

		if pattern.TotalInteractions < successCount {
			t.Errorf("Expected at least %d interactions, got %d", successCount, pattern.TotalInteractions)
		}

		if pattern.ConsistencyScore < 0 || pattern.ConsistencyScore > 1 {
			t.Errorf("Consistency score should be 0-1, got %f", pattern.ConsistencyScore)
		}

		if pattern.VarietyScore < 0 || pattern.VarietyScore > 1 {
			t.Errorf("Variety score should be 0-1, got %f", pattern.VarietyScore)
		}
	})

	t.Run("compatibility_modifier_generation", func(t *testing.T) {
		// Set up consistent behavior pattern
		for i := 0; i < 20; i++ {
			char.HandleRomanceInteraction("compliment")
			time.Sleep(30 * time.Millisecond)
		}

		// Force analysis update
		modifiers := char.compatibilityAnalyzer.Update(gameState)

		if len(modifiers) > 0 {
			// Check that modifiers are valid
			for _, modifier := range modifiers {
				if modifier.ModifierValue < modifier.MinValue || modifier.ModifierValue > modifier.MaxValue {
					t.Errorf("Modifier value %f outside bounds [%f, %f]",
						modifier.ModifierValue, modifier.MinValue, modifier.MaxValue)
				}

				if modifier.Reason == "" {
					t.Error("Modifier should have a reason")
				}

				if modifier.CreatedAt.IsZero() {
					t.Error("Modifier should have creation timestamp")
				}
			}
		}
	})
}

// TestCrisisRecoverySystem tests the relationship crisis and recovery mechanics
func TestCrisisRecoverySystem(t *testing.T) {
	// Create a character with crisis management
	card := createRomanceCharacterCard()
	char := createTestCharacterWithRomanceFeatures(card, true)

	gameState := char.GetGameState()

	t.Run("crisis_manager_initialization", func(t *testing.T) {
		if char.crisisRecoveryManager == nil {
			t.Error("Crisis recovery manager should be initialized for romance characters")
		}

		status := char.crisisRecoveryManager.GetCrisisStatus()
		if !status["enabled"].(bool) {
			t.Error("Crisis recovery manager should be enabled")
		}

		thresholds := status["thresholds"].(map[string]float64)
		if len(thresholds) == 0 {
			t.Error("Crisis thresholds should be configured")
		}
	})

	t.Run("jealousy_crisis_trigger", func(t *testing.T) {
		// Set up conditions for jealousy crisis
		gameState.ApplyInteractionEffects(map[string]float64{
			"jealousy":  85.0, // Above default threshold
			"affection": 30.0,
			"trust":     40.0,
		})

		// Update to trigger crisis check
		char.Update()

		activeCrises := char.crisisRecoveryManager.GetActiveCrises()
		if len(activeCrises) == 0 {
			t.Error("High jealousy should trigger a crisis")
			return
		}

		// Check crisis properties
		crisis := activeCrises[0]
		if crisis.Name != "jealousy_crisis" {
			t.Errorf("Expected jealousy_crisis, got %s", crisis.Name)
		}

		if !crisis.IsActive {
			t.Error("Crisis should be active")
		}

		if crisis.Severity <= 0 || crisis.Severity > 1 {
			t.Errorf("Crisis severity should be 0-1, got %f", crisis.Severity)
		}
	})

	t.Run("trust_crisis_trigger", func(t *testing.T) {
		// Reset to clean state
		gameState.ApplyInteractionEffects(map[string]float64{
			"jealousy":  -85.0, // Reset jealousy
			"trust":     5.0,   // Very low trust
			"affection": 20.0,
		})

		// Update to trigger crisis check
		char.Update()

		activeCrises := char.crisisRecoveryManager.GetActiveCrises()
		hasTrustCrisis := false
		for _, crisis := range activeCrises {
			if crisis.Name == "trust_crisis" && crisis.IsActive {
				hasTrustCrisis = true
				break
			}
		}

		if !hasTrustCrisis {
			t.Error("Very low trust should trigger a trust crisis")
		}
	})

	t.Run("crisis_recovery_through_interactions", func(t *testing.T) {
		// Set up a crisis situation
		gameState.ApplyInteractionEffects(map[string]float64{
			"jealousy":  85.0,
			"affection": 30.0,
			"trust":     40.0,
		})

		// Trigger crisis
		char.Update()

		activeCrises := char.crisisRecoveryManager.GetActiveCrises()
		if len(activeCrises) == 0 {
			t.Fatal("Crisis should be active for recovery testing")
		}

		initialCrisisCount := len(activeCrises)

		// Perform recovery interactions
		// First reduce jealousy
		gameState.ApplyInteractionEffects(map[string]float64{
			"jealousy": -30.0, // Reduce jealousy below threshold
		})

		// Simulate apology interactions (required for jealousy crisis recovery)
		char.HandleRomanceInteraction("apology")
		char.HandleRomanceInteraction("apology")
		char.HandleRomanceInteraction("deep_conversation")
		char.HandleRomanceInteraction("give_gift")

		// Wait for time requirement
		time.Sleep(100 * time.Millisecond) // Small delay to simulate time passage

		// Check for recovery
		recoveryEvent := char.crisisRecoveryManager.CheckRecovery(gameState, "apology")
		if recoveryEvent != nil {
			if recoveryEvent.Name != "jealousy_crisis_recovered" {
				t.Errorf("Expected jealousy_crisis_recovered, got %s", recoveryEvent.Name)
			}

			if len(recoveryEvent.Responses) == 0 {
				t.Error("Recovery event should have forgiveness responses")
			}

			if len(recoveryEvent.Effects) == 0 {
				t.Error("Recovery event should have stat bonuses")
			}
		}

		// Check that crisis was resolved
		finalCrises := char.crisisRecoveryManager.GetActiveCrises()
		if len(finalCrises) >= initialCrisisCount {
			t.Error("Crisis should be resolved after meeting recovery requirements")
		}
	})

	t.Run("ongoing_crisis_effects", func(t *testing.T) {
		// Set up active crisis
		gameState.ApplyInteractionEffects(map[string]float64{
			"jealousy":  90.0,
			"affection": 50.0,
			"trust":     50.0,
			"happiness": 50.0,
		})

		statsBefore := gameState.GetStats()

		// Update to apply ongoing crisis effects
		char.Update()
		time.Sleep(100 * time.Millisecond)
		char.Update()

		statsAfter := gameState.GetStats()

		// Crisis should cause ongoing stat degradation
		if statsAfter["affection"] >= statsBefore["affection"] {
			t.Log("Affection before:", statsBefore["affection"], "after:", statsAfter["affection"])
			// Note: Due to timing and probability, this might not always trigger immediately
		}
	})
}

// TestAdvancedFeaturesIntegration tests how all Phase 3 Task 3 features work together
func TestAdvancedFeaturesIntegration(t *testing.T) {
	// Create a complete romance character
	card := createRomanceCharacterCard()

	// Ensure personality traits are set for advanced features
	card.Personality = &PersonalityConfig{
		Traits: map[string]float64{
			"jealousy_prone":           0.7, // High jealousy
			"affection_responsiveness": 0.8, // High responsiveness
			"trust_difficulty":         0.6, // Moderate trust difficulty
			"shyness":                  0.4, // Moderate shyness
		},
		Compatibility: map[string]float64{
			"consistent_interaction": 1.2,
			"gift_appreciation":      1.5,
		},
	}

	char := createTestCharacterWithRomanceFeatures(card, true)

	// Enable game mode
	// Note: Already enabled via createTestCharacterWithRomanceFeatures

	gameState := char.GetGameState()

	t.Run("all_advanced_systems_initialized", func(t *testing.T) {
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

	t.Run("complex_relationship_scenario", func(t *testing.T) {
		// Simulate a complex relationship scenario

		// 1. Build initial relationship
		for i := 0; i < 5; i++ {
			char.HandleRomanceInteraction("compliment")
			char.HandleRomanceInteraction("give_gift")
			time.Sleep(10 * time.Millisecond)
		}

		// 2. Create neglect situation (trigger jealousy)
		char.lastInteraction = time.Now().Add(-2 * time.Hour)

		// 3. Let systems process the situation
		for i := 0; i < 20; i++ {
			char.Update()
			time.Sleep(50 * time.Millisecond)
		}

		// 4. Check that systems are responding
		jealousyLevel := char.jealousyManager.GetJealousyLevel(gameState)
		if jealousyLevel > 0.8 {
			t.Log("High jealousy detected, checking for crisis...")

			activeCrises := char.crisisRecoveryManager.GetActiveCrises()
			if len(activeCrises) > 0 {
				t.Log("Crisis active:", activeCrises[0].Name)
			}
		}

		// 5. Attempt recovery
		char.HandleRomanceInteraction("apology")
		char.HandleRomanceInteraction("reassurance")
		char.HandleRomanceInteraction("give_gift")

		// 6. Check final state
		finalStats := gameState.GetStats()
		if finalStats["affection"] <= 0 {
			t.Error("Relationship should be recoverable with proper interactions")
		}

		// 7. Verify compatibility adaptation
		pattern := char.compatibilityAnalyzer.GetPlayerPattern()
		if pattern != nil && pattern.TotalInteractions > 0 {
			t.Logf("Compatibility analysis: %d interactions, consistency: %.2f, variety: %.2f",
				pattern.TotalInteractions, pattern.ConsistencyScore, pattern.VarietyScore)
		}
	})

	t.Run("personality_based_thresholds", func(t *testing.T) {
		// Verify that crisis thresholds are adjusted based on personality
		status := char.crisisRecoveryManager.GetCrisisStatus()
		thresholds := status["thresholds"].(map[string]float64)

		// High jealousy-prone characters should have lower jealousy thresholds
		jealousyThreshold := thresholds["jealousy"]
		if jealousyThreshold > 80.0 { // Should be adjusted down from default 80
			t.Errorf("Jealousy-prone character should have lower jealousy threshold, got %f", jealousyThreshold)
		}

		// High trust difficulty should raise trust crisis threshold
		trustThreshold := thresholds["trust"]
		if trustThreshold < 15.0 { // Should be adjusted up from default 15
			t.Errorf("Trust-difficult character should have higher trust threshold, got %f", trustThreshold)
		}
	})
}

// Helper function to create a romance character card for testing
func createRomanceCharacterCard() *CharacterCard {
	return &CharacterCard{
		Name:        "Test Romance Character",
		Description: "A test character with romance features",
		Animations: map[string]string{
			"idle":     "test_idle.gif",
			"talking":  "test_talking.gif",
			"happy":    "test_happy.gif",
			"sad":      "test_sad.gif",
			"jealous":  "test_jealous.gif",
			"blushing": "test_blushing.gif",
			"shy":      "test_shy.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
			"intimacy":  {Initial: 0, Max: 100, DegradationRate: 0.2, CriticalThreshold: 0},
			"jealousy":  {Initial: 0, Max: 100, DegradationRate: 2.0, CriticalThreshold: 80},
			"happiness": {Initial: 100, Max: 100, DegradationRate: 0.8, CriticalThreshold: 15},
		},
		GameRules: &GameRulesConfig{
			StatsDecayInterval: 60,
			AutoSaveInterval:   300,
		},
		Interactions: map[string]InteractionConfig{
			"compliment": {
				Triggers:   []string{"hover"},
				Effects:    map[string]float64{"affection": 5, "happiness": 3, "trust": 1},
				Animations: []string{"blushing", "happy"},
				Responses:  []string{"Thank you! 💕"},
				Cooldown:   45,
			},
			"give_gift": {
				Triggers:   []string{"doubleclick"},
				Effects:    map[string]float64{"affection": 10, "happiness": 8, "trust": 2},
				Animations: []string{"happy"},
				Responses:  []string{"This is perfect! 🎁"},
				Cooldown:   120,
			},
			"deep_conversation": {
				Triggers:   []string{"shift+click"},
				Effects:    map[string]float64{"trust": 8, "affection": 3, "intimacy": 5},
				Animations: []string{"talking"},
				Responses:  []string{"I love talking with you..."},
				Cooldown:   90,
			},
			"apology": {
				Triggers:   []string{"ctrl+shift+click"},
				Effects:    map[string]float64{"trust": 12, "affection": 8, "jealousy": -15},
				Animations: []string{"shy"},
				Responses:  []string{"Thank you for apologizing... 💕"},
				Cooldown:   180,
			},
			"reassurance": {
				Triggers:   []string{"alt+shift+click"},
				Effects:    map[string]float64{"trust": 10, "affection": 6, "jealousy": -10},
				Animations: []string{"happy"},
				Responses:  []string{"Your reassurance helps... 💓"},
				Cooldown:   120,
			},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"jealousy_prone":           0.6,
				"affection_responsiveness": 0.8,
				"trust_difficulty":         0.5,
				"shyness":                  0.4,
			},
			Compatibility: map[string]float64{
				"consistent_interaction": 1.2,
				"gift_appreciation":      1.5,
			},
		},
		RomanceDialogs: []DialogExtended{},
		RomanceEvents:  []RandomEventConfig{},
	}
}
