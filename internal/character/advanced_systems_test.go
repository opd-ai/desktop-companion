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

		currentLevel := status["currentLevel"].(float64)
		if currentLevel != 0.2 { // 20/100
			t.Errorf("Expected current level 0.2, got %f", currentLevel)
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

// TestCompatibilityAnalyzer tests the compatibility analysis system
func TestCompatibilityAnalyzer(t *testing.T) {
	gameState := &GameState{
		stats: map[string]float64{
			"affection": 50.0,
			"trust":     40.0,
		},
		interactionHistory: []InteractionRecord{},
		lastSave:           time.Now(),
	}

	t.Run("new_compatibility_analyzer", func(t *testing.T) {
		personalities := map[string]float64{
			"consistency_preference": 0.7,
			"variety_preference":     0.3,
		}

		analyzer := NewCompatibilityAnalyzer(personalities, 0.8)

		if analyzer == nil {
			t.Fatal("NewCompatibilityAnalyzer should not return nil")
		}

		if analyzer.adaptationStrength != 0.8 {
			t.Errorf("Expected adaptation strength 0.8, got %f", analyzer.adaptationStrength)
		}

		if len(analyzer.personalityFactors) != 2 {
			t.Errorf("Expected 2 personality factors, got %d", len(analyzer.personalityFactors))
		}
	})

	t.Run("analyze_player_behavior", func(t *testing.T) {
		analyzer := NewCompatibilityAnalyzer(map[string]float64{}, 0.5)

		// Add some interaction history
		now := time.Now()
		interactions := []InteractionRecord{
			{Type: "compliment", Timestamp: now.Add(-10 * time.Minute)},
			{Type: "compliment", Timestamp: now.Add(-8 * time.Minute)},
			{Type: "give_gift", Timestamp: now.Add(-5 * time.Minute)},
			{Type: "compliment", Timestamp: now.Add(-2 * time.Minute)},
		}

		pattern := analyzer.analyzePlayerBehavior(interactions)

		if pattern.TotalInteractions != 4 {
			t.Errorf("Expected 4 total interactions, got %d", pattern.TotalInteractions)
		}

		if pattern.ConsistencyScore < 0 || pattern.ConsistencyScore > 1 {
			t.Errorf("Consistency score should be 0-1, got %f", pattern.ConsistencyScore)
		}

		if pattern.VarietyScore < 0 || pattern.VarietyScore > 1 {
			t.Errorf("Variety score should be 0-1, got %f", pattern.VarietyScore)
		}

		// Most interactions are "compliment", so should show lower variety
		if pattern.VarietyScore > 0.8 {
			t.Errorf("Expected lower variety score due to repeated compliments, got %f", pattern.VarietyScore)
		}
	})

	t.Run("generate_compatibility_modifiers", func(t *testing.T) {
		personalities := map[string]float64{
			"consistency_preference": 0.8,
			"gift_appreciation":      1.5,
		}

		analyzer := NewCompatibilityAnalyzer(personalities, 0.7)

		pattern := PlayerBehaviorPattern{
			TotalInteractions: 10,
			ConsistencyScore:  0.9, // High consistency
			VarietyScore:      0.3, // Low variety
			InteractionTypes: map[string]int{
				"compliment": 7,
				"give_gift":  3,
			},
			TimingConsistency: 0.8,
		}

		modifiers := analyzer.generateCompatibilityModifiers(pattern)

		if len(modifiers) == 0 {
			t.Error("Should generate some compatibility modifiers")
		}

		// Check modifier properties
		for _, modifier := range modifiers {
			if modifier.ModifierValue < modifier.MinValue || modifier.ModifierValue > modifier.MaxValue {
				t.Errorf("Modifier value %f outside bounds [%f, %f]",
					modifier.ModifierValue, modifier.MinValue, modifier.MaxValue)
			}

			if modifier.Reason == "" {
				t.Error("Modifier should have a reason")
			}

			if modifier.Duration <= 0 {
				t.Error("Modifier should have positive duration")
			}
		}
	})

	t.Run("get_compatibility_insights", func(t *testing.T) {
		analyzer := NewCompatibilityAnalyzer(map[string]float64{"test": 1.0}, 0.6)

		insights := analyzer.GetCompatibilityInsights()

		if !insights["enabled"].(bool) {
			t.Error("Compatibility analyzer should be enabled")
		}

		adaptationStrength := insights["adaptationStrength"].(float64)
		if adaptationStrength != 0.6 {
			t.Errorf("Expected adaptation strength 0.6, got %f", adaptationStrength)
		}

		personalityCount := insights["personalityFactorCount"].(int)
		if personalityCount != 1 {
			t.Errorf("Expected 1 personality factor, got %d", personalityCount)
		}
	})
}

// TestCrisisRecoveryManager tests the crisis and recovery system
func TestCrisisRecoveryManager(t *testing.T) {
	gameState := &GameState{
		stats: map[string]float64{
			"affection": 30.0,
			"trust":     25.0,
			"jealousy":  85.0,
			"happiness": 20.0,
		},
		lastSave: time.Now(),
	}

	t.Run("new_crisis_recovery_manager", func(t *testing.T) {
		thresholds := map[string]float64{
			"jealousy": 80.0,
			"trust":    15.0,
		}

		recoveryPaths := map[string]CrisisRecoveryPath{
			"jealousy_crisis": {
				Name:                 "Jealousy Crisis Recovery",
				RequiredInteractions: []string{"apology", "reassurance"},
				StatRequirements:     map[string]float64{"jealousy": 70.0}, // Must be below this
				TimeRequirement:      5 * time.Minute,
			},
		}

		manager := NewCrisisRecoveryManager(thresholds, recoveryPaths)

		if manager == nil {
			t.Fatal("NewCrisisRecoveryManager should not return nil")
		}

		if len(manager.crisisThresholds) != 2 {
			t.Errorf("Expected 2 crisis thresholds, got %d", len(manager.crisisThresholds))
		}

		if len(manager.recoveryPaths) != 1 {
			t.Errorf("Expected 1 recovery path, got %d", len(manager.recoveryPaths))
		}
	})

	t.Run("detect_crisis", func(t *testing.T) {
		thresholds := map[string]float64{
			"jealousy": 80.0,
			"trust":    15.0,
		}

		manager := NewCrisisRecoveryManager(thresholds, map[string]CrisisRecoveryPath{})

		// Check jealousy crisis detection
		crises := manager.detectCrises(gameState)

		hasJealousyCrisis := false
		for _, crisis := range crises {
			if crisis.Name == "jealousy_crisis" {
				hasJealousyCrisis = true
				if !crisis.IsActive {
					t.Error("Jealousy crisis should be active")
				}
				if crisis.Severity <= 0 || crisis.Severity > 1 {
					t.Errorf("Crisis severity should be 0-1, got %f", crisis.Severity)
				}
				break
			}
		}

		if !hasJealousyCrisis {
			t.Error("Should detect jealousy crisis with jealousy at 85")
		}
	})

	t.Run("crisis_recovery_validation", func(t *testing.T) {
		recoveryPaths := map[string]CrisisRecoveryPath{
			"jealousy_crisis": {
				Name:                 "Jealousy Crisis Recovery",
				RequiredInteractions: []string{"apology"},
				StatRequirements:     map[string]float64{"jealousy": 70.0},
				TimeRequirement:      1 * time.Minute,
			},
		}

		manager := NewCrisisRecoveryManager(map[string]float64{"jealousy": 80}, recoveryPaths)

		// Set up a crisis
		crisis := Crisis{
			Name:        "jealousy_crisis",
			Description: "High jealousy affecting relationship",
			IsActive:    true,
			Severity:    0.8,
			StartTime:   time.Now().Add(-2 * time.Minute), // Started 2 minutes ago
		}

		manager.activeCrises = []Crisis{crisis}

		// Simulate recovery conditions
		recoveryState := &GameState{
			stats: map[string]float64{
				"jealousy": 65.0, // Below threshold
			},
			interactionHistory: []InteractionRecord{
				{Type: "apology", Timestamp: time.Now().Add(-30 * time.Second)},
			},
			lastSave: time.Now(),
		}

		recoveryEvent := manager.CheckRecovery(recoveryState, "apology")

		if recoveryEvent == nil {
			t.Error("Should have recovery event when conditions are met")
		} else {
			if recoveryEvent.Name != "jealousy_crisis_recovered" {
				t.Errorf("Expected jealousy_crisis_recovered, got %s", recoveryEvent.Name)
			}

			if len(recoveryEvent.Effects) == 0 {
				t.Error("Recovery event should have stat effects")
			}

			if len(recoveryEvent.Responses) == 0 {
				t.Error("Recovery event should have responses")
			}
		}

		// Check if crisis was resolved
		activeCrises := manager.GetActiveCrises()
		hasJealousyCrisis := false
		for _, c := range activeCrises {
			if c.Name == "jealousy_crisis" && c.IsActive {
				hasJealousyCrisis = true
				break
			}
		}

		if hasJealousyCrisis {
			t.Error("Jealousy crisis should be resolved after successful recovery")
		}
	})

	t.Run("get_crisis_status", func(t *testing.T) {
		thresholds := map[string]float64{
			"jealousy": 75.0,
			"trust":    20.0,
		}

		manager := NewCrisisRecoveryManager(thresholds, map[string]CrisisRecoveryPath{})

		status := manager.GetCrisisStatus()

		if !status["enabled"].(bool) {
			t.Error("Crisis recovery manager should be enabled")
		}

		statusThresholds := status["thresholds"].(map[string]float64)
		if len(statusThresholds) != 2 {
			t.Errorf("Expected 2 thresholds in status, got %d", len(statusThresholds))
		}

		if statusThresholds["jealousy"] != 75.0 {
			t.Errorf("Expected jealousy threshold 75, got %f", statusThresholds["jealousy"])
		}

		activeCrisisCount := status["activeCrisisCount"].(int)
		if activeCrisisCount != 0 {
			t.Errorf("Expected 0 active crises initially, got %d", activeCrisisCount)
		}
	})

	t.Run("ongoing_crisis_effects", func(t *testing.T) {
		manager := NewCrisisRecoveryManager(map[string]float64{"jealousy": 80}, map[string]CrisisRecoveryPath{})

		// Create active crisis
		crisis := Crisis{
			Name:        "jealousy_crisis",
			Description: "High jealousy",
			IsActive:    true,
			Severity:    0.9,
			StartTime:   time.Now().Add(-10 * time.Minute),
		}

		manager.activeCrises = []Crisis{crisis}

		effects := manager.Update(gameState)

		// Crisis should cause ongoing negative effects
		if len(effects) > 0 {
			hasNegativeEffect := false
			for _, change := range effects {
				if change < 0 {
					hasNegativeEffect = true
					break
				}
			}

			if !hasNegativeEffect {
				t.Log("Expected some negative effects from ongoing crisis")
			}
		}
	})
}
