package character

import (
	"testing"
	"time"
)

// TestEnhancedDialogueSystem tests the relationship-aware dialogue selection from Phase 2
func TestEnhancedDialogueSystem(t *testing.T) {
	// Create a character card with romance dialogs
	card := &CharacterCard{
		Name:        "Enhanced Dialogue Test Character",
		Description: "Test character for enhanced dialogue system",
		Animations: map[string]string{
			"idle":          "idle.gif",
			"talking":       "talking.gif",
			"romantic_idle": "romantic_idle.gif",
			"blushing":      "blushing.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello there!"}, Animation: "talking", Cooldown: 5},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
			"trust":     {Initial: 10, Max: 100, DegradationRate: 0.05, CriticalThreshold: 5},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shyness":     0.6,
				"romanticism": 0.8,
				"flirtiness":  0.4,
			},
		},
		RomanceDialogs: []DialogExtended{
			{
				Dialog: Dialog{
					Trigger:   "click",
					Responses: []string{"Hi sweetheart! 💕", "I was hoping you'd come see me!"},
					Animation: "romantic_idle",
					Cooldown:  5,
				},
				Requirements: &RomanceRequirement{
					Stats: map[string]map[string]float64{
						"affection": {"min": 20},
					},
				},
			},
			{
				Dialog: Dialog{
					Trigger:   "hover",
					Responses: []string{"*heart flutters* 💓", "Just being near you makes me happy..."},
					Animation: "blushing",
					Cooldown:  10,
				},
				Requirements: &RomanceRequirement{
					Stats: map[string]map[string]float64{
						"affection": {"min": 30},
						"trust":     {"min": 20},
					},
				},
			},
		},
	}

	// Test with low affection - should use regular dialog
	t.Run("low affection uses regular dialog", func(t *testing.T) {
		char := createTestCharacterInstance(card, true)

		// Ensure affection is low
		char.gameState.Stats["affection"].Current = 10

		response := char.HandleClick()
		if response != "Hello there!" {
			t.Errorf("Expected regular dialog with low affection, got: %s", response)
		}
	})

	// Test with high affection - should use romance dialog
	t.Run("high affection uses romance dialog", func(t *testing.T) {
		char := createTestCharacterInstance(card, true)

		// Set high affection
		char.gameState.Stats["affection"].Current = 25

		response := char.HandleClick()
		expectedResponses := []string{"Hi sweetheart! 💕", "I was hoping you'd come see me!"}

		if !containsDialog(expectedResponses, response) {
			t.Errorf("Expected romance dialog with high affection, got: %s", response)
		}
	})

	// Test hover dialog requirements
	t.Run("hover dialog requires both affection and trust", func(t *testing.T) {
		char := createTestCharacterInstance(card, true)

		// Set high affection but low trust
		char.gameState.Stats["affection"].Current = 35
		char.gameState.Stats["trust"].Current = 15

		// Reset lastInteraction to avoid the 2-second guard in HandleHover
		char.lastInteraction = time.Now().Add(-5 * time.Second)

		response := char.HandleHover()
		if response != "" {
			t.Errorf("Expected no hover dialog with insufficient trust, got: %s", response)
		}

		// Set sufficient trust
		char.gameState.Stats["trust"].Current = 25

		// Reset lastInteraction again to avoid the guard
		char.lastInteraction = time.Now().Add(-5 * time.Second)

		response = char.HandleHover()
		expectedResponses := []string{"*heart flutters* 💓", "Just being near you makes me happy..."}

		if !containsDialog(expectedResponses, response) {
			t.Errorf("Expected hover dialog with sufficient stats, got: %s", response)
		}
	})
}

// TestRomanceDialogCooldowns tests that romance dialogs respect cooldowns
func TestRomanceDialogCooldowns(t *testing.T) {
	card := &CharacterCard{
		Name:        "Cooldown Test Character",
		Description: "Test character for dialog cooldowns",
		Animations: map[string]string{
			"idle":          "idle.gif",
			"romantic_idle": "romantic_idle.gif",
		},
		Dialogs:  []Dialog{},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 50, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{"romanticism": 0.8},
		},
		RomanceDialogs: []DialogExtended{
			{
				Dialog: Dialog{
					Trigger:   "click",
					Responses: []string{"Hello love!"},
					Animation: "romantic_idle",
					Cooldown:  2, // 2-second cooldown
				},
				Requirements: &RomanceRequirement{
					Stats: map[string]map[string]float64{
						"affection": {"min": 20},
					},
				},
			},
		},
	}

	char := createTestCharacterInstance(card, true)

	// First click should work
	response1 := char.HandleClick()
	if response1 != "Hello love!" {
		t.Errorf("First click should return romance dialog, got: %s", response1)
	}

	// Immediate second click should be blocked by cooldown
	response2 := char.HandleClick()
	if response2 != "" {
		t.Errorf("Second immediate click should be blocked by cooldown, got: %s", response2)
	}

	// Wait for cooldown to expire
	time.Sleep(3 * time.Second)
	response3 := char.HandleClick()
	if response3 != "Hello love!" {
		t.Errorf("Click after cooldown should work again, got: %s", response3)
	}
}

// TestDialogScoring tests the personality-based dialog scoring system
func TestDialogScoring(t *testing.T) {
	card := &CharacterCard{
		Name:        "Scoring Test Character",
		Description: "Test character for dialog scoring",
		Animations: map[string]string{
			"idle":     "idle.gif",
			"romantic": "romantic.gif",
		},
		Dialogs:  []Dialog{},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 60, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shyness":     0.8, // Very shy
				"romanticism": 0.9, // Very romantic
				"flirtiness":  0.2, // Not flirty
			},
		},
		RomanceDialogs: []DialogExtended{
			{
				Dialog: Dialog{
					Trigger:   "click",
					Responses: []string{"*boldly* Hey there, gorgeous! 😘"}, // Flirty response
					Animation: "romantic",
					Cooldown:  5,
				},
				Requirements: &RomanceRequirement{
					Stats: map[string]map[string]float64{
						"affection": {"min": 20},
					},
				},
			},
			{
				Dialog: Dialog{
					Trigger:   "click",
					Responses: []string{"Hi... *blushes softly*"}, // Shy response
					Animation: "romantic",
					Cooldown:  5,
				},
				Requirements: &RomanceRequirement{
					Stats: map[string]map[string]float64{
						"affection": {"min": 20},
					},
				},
			},
		},
	}

	char := createTestCharacterInstance(card, true)

	// Test multiple times to see if shy character prefers shy responses
	shyResponseCount := 0
	boldResponseCount := 0

	for i := 0; i < 10; i++ {
		// Reset cooldown
		char.dialogCooldowns = make(map[string]time.Time)

		response := char.HandleClick()
		if response == "Hi... *blushes softly*" {
			shyResponseCount++
		} else if response == "*boldly* Hey there, gorgeous! 😘" {
			boldResponseCount++
		}
	}

	t.Logf("Shy character responses: %d shy, %d bold", shyResponseCount, boldResponseCount)

	// Shy character should prefer shy responses more often
	if shyResponseCount < boldResponseCount {
		t.Errorf("Shy character should prefer shy responses, got %d shy vs %d bold", shyResponseCount, boldResponseCount)
	}
}

// TestNoRomanceDialogs tests behavior when RomanceDialogs is nil or empty
func TestNoRomanceDialogs(t *testing.T) {
	card := &CharacterCard{
		Name:        "No Romance Test Character",
		Description: "Test character without romance dialogs",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"talking": "talking.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Regular hello!"}, Animation: "talking", Cooldown: 5},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 50, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{"romanticism": 0.8},
		},
		// RomanceDialogs is nil
	}

	char := createTestCharacterInstance(card, true)

	response := char.HandleClick()
	if response != "Regular hello!" {
		t.Errorf("Character without romance dialogs should use regular dialog, got: %s", response)
	}
}

// TestInteractionCountRequirements tests dialog unlocking based on interaction counts
func TestInteractionCountRequirements(t *testing.T) {
	card := &CharacterCard{
		Name:        "Interaction Count Test Character",
		Description: "Test character for interaction count requirements",
		Animations: map[string]string{
			"idle":    "idle.gif",
			"special": "special.gif",
		},
		Dialogs:  []Dialog{},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 50, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{"romanticism": 0.8},
		},
		Progression: &ProgressionConfig{
			Levels:       []LevelConfig{},
			Achievements: []AchievementConfig{},
		},
		RomanceDialogs: []DialogExtended{
			{
				Dialog: Dialog{
					Trigger:   "click",
					Responses: []string{"We've become so close!"},
					Animation: "special",
					Cooldown:  5,
				},
				Requirements: &RomanceRequirement{
					InteractionCount: map[string]map[string]int{
						"compliment": {"min": 3}, // Requires 3+ compliments
					},
				},
			},
		},
	}

	char := createTestCharacterInstance(card, true)

	// Initially, no interactions recorded - dialog should not be available
	response := char.HandleClick()
	if response != "" {
		t.Errorf("Dialog should not be available without sufficient interactions, got: %s", response)
	}

	// Record some compliment interactions
	if char.gameState.Progression != nil {
		char.gameState.Progression.RecordInteraction("compliment")
		char.gameState.Progression.RecordInteraction("compliment")
		char.gameState.Progression.RecordInteraction("compliment")
	}

	// Now dialog should be available
	response = char.HandleClick()
	if response != "We've become so close!" {
		t.Errorf("Dialog should be available after sufficient interactions, got: %s", response)
	}
}

// Helper function to check if a slice contains a specific string
func containsDialog(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
