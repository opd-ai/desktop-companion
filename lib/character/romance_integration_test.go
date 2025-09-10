package character

import (
	"testing"
)

// TestBackwardCompatibility ensures that existing characters without romance features continue to work
func TestBackwardCompatibility(t *testing.T) {
	// Test basic character card (similar to existing default character)
	basicCard := CharacterCard{
		Name:        "Basic Pet",
		Description: "A simple desktop pet",
		Animations: map[string]string{
			"idle":    "animations/idle.gif",
			"talking": "animations/talking.gif",
			"happy":   "animations/happy.gif",
		},
		Dialogs: []Dialog{
			{
				Trigger:   "click",
				Responses: []string{"Hello!", "How are you?"},
				Animation: "talking",
				Cooldown:  5,
			},
		},
		Behavior: Behavior{
			IdleTimeout:     30,
			MovementEnabled: true,
			DefaultSize:     128,
		},
	}

	// Should validate successfully without romance features
	err := basicCard.Validate()
	if err != nil {
		t.Errorf("Basic character validation failed: %v", err)
	}

	// Should not have romance features
	if basicCard.HasRomanceFeatures() {
		t.Error("Basic character should not have romance features")
	}

	// Should return default personality values
	shyness := basicCard.GetPersonalityTrait("shyness")
	if shyness != 0.5 {
		t.Errorf("Expected default shyness 0.5, got %f", shyness)
	}

	giftMod := basicCard.GetCompatibilityModifier("gift_appreciation")
	if giftMod != 1.0 {
		t.Errorf("Expected default gift modifier 1.0, got %f", giftMod)
	}
}

// TestGameFeaturesWithRomance ensures game features and romance features work together
func TestGameFeaturesWithRomance(t *testing.T) {
	hybridCard := CharacterCard{
		Name:        "Hybrid Character",
		Description: "Character with both game and romance features",
		Animations: map[string]string{
			"idle":     "animations/idle.gif",
			"talking":  "animations/talking.gif",
			"blushing": "animations/blushing.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello"}, Animation: "talking"},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		// Traditional game stats
		Stats: map[string]StatConfig{
			"hunger":    {Initial: 100, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
			"happiness": {Initial: 100, Max: 100, DegradationRate: 0.8, CriticalThreshold: 15},
			// Romance stats
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		},
		GameRules: &GameRulesConfig{
			StatsDecayInterval:             60,
			AutoSaveInterval:               300,
			CriticalStateAnimationPriority: true,
			DeathEnabled:                   false,
			EvolutionEnabled:               true,
			MoodBasedAnimations:            true,
		},
		// Traditional game interactions
		Interactions: map[string]InteractionConfig{
			"feed": {
				Triggers:  []string{"rightclick"},
				Effects:   map[string]float64{"hunger": 25, "happiness": 5},
				Responses: []string{"Yum!"},
				Cooldown:  30,
			},
			// Romance interactions
			"compliment": {
				Triggers:  []string{"shift+click"},
				Effects:   map[string]float64{"affection": 5, "happiness": 3},
				Responses: []string{"Thank you!"},
				Cooldown:  45,
			},
		},
		// Romance features
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shyness":     0.6,
				"romanticism": 0.8,
			},
		},
	}

	// Should validate successfully with both game and romance features
	err := hybridCard.Validate()
	if err != nil {
		t.Errorf("Hybrid character validation failed: %v", err)
	}

	// Should have both game and romance features
	if !hybridCard.HasGameFeatures() {
		t.Error("Hybrid character should have game features")
	}

	if !hybridCard.HasRomanceFeatures() {
		t.Error("Hybrid character should have romance features")
	}
}

// TestRomanceStatsIntegration tests that romance stats work with the existing stats system
func TestRomanceStatsIntegration(t *testing.T) {
	// Create stats configuration that includes romance stats
	statConfigs := map[string]StatConfig{
		"hunger":    {Initial: 100, Max: 100, DegradationRate: 1.0, CriticalThreshold: 20},
		"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		"intimacy":  {Initial: 0, Max: 100, DegradationRate: 0.2, CriticalThreshold: 0},
	}

	gameConfig := &GameConfig{
		StatsDecayInterval:             60,
		CriticalStateAnimationPriority: true,
		MoodBasedAnimations:            true,
	}

	// Create game state with romance stats
	gameState := NewGameState(statConfigs, gameConfig)

	// Test that romance stats are properly initialized
	affection := gameState.GetStat("affection")
	if affection != 0 {
		t.Errorf("Expected initial affection 0, got %f", affection)
	}

	trust := gameState.GetStat("trust")
	if trust != 20 {
		t.Errorf("Expected initial trust 20, got %f", trust)
	}

	// Test applying romance interaction effects
	romanceEffects := map[string]float64{
		"affection": 5,
		"trust":     2,
		"happiness": 3, // Traditional stat
	}

	gameState.ApplyInteractionEffects(romanceEffects)

	// Verify effects were applied
	newAffection := gameState.GetStat("affection")
	if newAffection != 5 {
		t.Errorf("Expected affection 5 after effect, got %f", newAffection)
	}

	newTrust := gameState.GetStat("trust")
	if newTrust != 22 {
		t.Errorf("Expected trust 22 after effect, got %f", newTrust)
	}

	// Test that romance stats respect boundaries
	extremeEffects := map[string]float64{
		"affection": 200, // Should be capped at max (100)
		"trust":     -50, // Should be capped at min (0)
	}

	gameState.ApplyInteractionEffects(extremeEffects)

	cappedAffection := gameState.GetStat("affection")
	if cappedAffection != 100 {
		t.Errorf("Expected affection capped at 100, got %f", cappedAffection)
	}

	cappedTrust := gameState.GetStat("trust")
	if cappedTrust < 0 {
		t.Errorf("Expected trust >= 0, got %f", cappedTrust)
	}
}

// TestRomanceProgressionIntegration tests that romance stats work with progression
func TestRomanceProgressionIntegration(t *testing.T) {
	// Test that progression requirements can include romance stats
	card := CharacterCard{
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 20, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		},
		Progression: &ProgressionConfig{
			Levels: []LevelConfig{
				{
					Name:        "Stranger",
					Requirement: map[string]int64{"age": 0},
					Size:        128,
				},
				{
					Name: "Friend",
					Requirement: map[string]int64{
						"age":       86400, // 1 day
						"affection": 15,
						"trust":     10,
					},
					Size: 132,
				},
				{
					Name: "Close Friend",
					Requirement: map[string]int64{
						"age":       172800, // 2 days
						"affection": 30,
						"trust":     25,
					},
					Size: 136,
				},
			},
		},
	}

	// Validate that progression with romance requirements works
	err := card.validateProgression()
	if err != nil {
		t.Errorf("Romance progression validation failed: %v", err)
	}
}
