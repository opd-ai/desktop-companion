package character

import (
	"testing"
	"time"
)

// TestHandleRomanceInteraction tests the HandleRomanceInteraction method
func TestHandleRomanceInteraction(t *testing.T) {
	// Create a test character card with romance features
	card := &CharacterCard{
		Name:        "Romance Test Character",
		Description: "Test character for romance interactions",
		Animations: map[string]string{
			"idle":     "idle.gif",
			"happy":    "happy.gif",
			"blushing": "blushing.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Hello!"}, Animation: "happy"},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 20, Max: 100, DegradationRate: 0.1},
			"trust":     {Initial: 15, Max: 100, DegradationRate: 0.05},
			"happiness": {Initial: 50, Max: 100, DegradationRate: 0.2},
		},
		Interactions: map[string]InteractionConfig{
			"compliment": {
				RomanceCategory: "verbal_affection",
				Responses:       []string{"That's so sweet!", "You think so?"},
				Animations:      []string{"blushing", "happy"},
				Effects:         map[string]float64{"affection": 5, "trust": 2},
				Cooldown:        10,
				Requirements:    map[string]map[string]float64{},
			},
			"give_gift": {
				RomanceCategory: "physical_gift",
				Responses:       []string{"Thank you!", "Beautiful!"},
				Animations:      []string{"happy"},
				Effects:         map[string]float64{"affection": 8, "happiness": 5},
				Cooldown:        30,
				Requirements:    map[string]map[string]float64{},
			},
			"non_romance_interaction": {
				// No romance category - should be ignored by HandleRomanceInteraction
				Responses:  []string{"Regular interaction"},
				Effects:    map[string]float64{"happiness": 1},
				Cooldown:   5,
				Requirements: map[string]map[string]float64{},
			},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{
				"shyness":                  0.6,
				"romanticism":              0.8,
				"affection_responsiveness": 1.2,
			},
			Compatibility: map[string]float64{
				"gift_appreciation":  1.5,
				"conversation_lover": 1.3,
			},
		},
	}

	tests := []struct {
		name              string
		interactionType   string
		expectResponse    bool
		expectEffects     bool
		gameFeatures      bool
		description       string
	}{
		{
			name:              "valid compliment interaction",
			interactionType:   "compliment",
			expectResponse:    true,
			expectEffects:     true,
			gameFeatures:      true,
			description:       "Should handle compliment interaction and return response",
		},
		{
			name:              "valid gift interaction",
			interactionType:   "give_gift",
			expectResponse:    true,
			expectEffects:     true,
			gameFeatures:      true,
			description:       "Should handle gift interaction and apply effects",
		},
		{
			name:              "non-existent interaction",
			interactionType:   "invalid_interaction",
			expectResponse:    false,
			expectEffects:     false,
			gameFeatures:      true,
			description:       "Should return empty response for non-existent interaction",
		},
		{
			name:              "non-romance interaction",
			interactionType:   "non_romance_interaction",
			expectResponse:    false,
			expectEffects:     false,
			gameFeatures:      true,
			description:       "Should ignore interactions without romance category",
		},
		{
			name:              "romance interaction without game features",
			interactionType:   "compliment",
			expectResponse:    false,
			expectEffects:     false,
			gameFeatures:      false,
			description:       "Should return empty when game features are disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create character instance
			char := createTestCharacterInstance(card, tt.gameFeatures)

			// Get initial affection stat if game features enabled
			var initialAffection float64
			if tt.gameFeatures && char.gameState != nil {
				if stat, exists := char.gameState.Stats["affection"]; exists {
					initialAffection = stat.Current
				}
			}

			// Test the interaction
			response := char.HandleRomanceInteraction(tt.interactionType)

			// Verify response expectation
			if tt.expectResponse && response == "" {
				t.Errorf("Expected response but got empty string")
			}
			if !tt.expectResponse && response != "" {
				t.Errorf("Expected empty response but got: %s", response)
			}

			// Verify effects were applied if expected
			if tt.expectEffects && tt.gameFeatures && char.gameState != nil {
				if stat, exists := char.gameState.Stats["affection"]; exists {
					if stat.Current == initialAffection {
						t.Errorf("Expected affection to change but it remained %f", initialAffection)
					}
				}
			}
		})
	}
}

// TestHandleRomanceInteractionCooldown tests that cooldowns work correctly
func TestHandleRomanceInteractionCooldown(t *testing.T) {
	card := &CharacterCard{
		Name:        "Cooldown Test Character",
		Description: "Test character for cooldown testing",
		Animations:  map[string]string{"idle": "idle.gif", "happy": "happy.gif"},
		Dialogs:     []Dialog{{Trigger: "click", Responses: []string{"Hello!"}}},
		Behavior:    Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 20, Max: 100, DegradationRate: 0.1},
		},
		Interactions: map[string]InteractionConfig{
			"compliment": {
				RomanceCategory: "verbal_affection",
				Responses:       []string{"Thank you!"},
				Effects:         map[string]float64{"affection": 5},
				Cooldown:        2, // 2 second cooldown
				Requirements:    map[string]map[string]float64{},
			},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{"affection_responsiveness": 1.0},
		},
	}

	char := createTestCharacterInstance(card, true)

	// First interaction should work
	response1 := char.HandleRomanceInteraction("compliment")
	if response1 == "" {
		t.Fatal("First interaction should return a response")
	}

	// Immediate second interaction should be blocked by cooldown
	response2 := char.HandleRomanceInteraction("compliment")
	if response2 != "" {
		t.Errorf("Second immediate interaction should be blocked by cooldown, but got: %s", response2)
	}

	// Wait for cooldown to expire and try again
	time.Sleep(2 * time.Second)
	response3 := char.HandleRomanceInteraction("compliment")
	if response3 == "" {
		t.Error("Interaction after cooldown should work again")
	}
}

// Helper function to create a test character instance
func createTestCharacterInstance(card *CharacterCard, gameFeatures bool) *Character {
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
	if gameFeatures && card.HasGameFeatures() {
		char.initializeGameFeatures()
	}

	return char
}