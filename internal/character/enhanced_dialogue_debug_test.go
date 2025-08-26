package character

import (
	"testing"
)

// TestRomanceDialogDebug helps debug the enhanced dialogue system
func TestRomanceDialogDebug(t *testing.T) {
	card := &CharacterCard{
		Name:        "Debug Character",
		Description: "Debug character for enhanced dialogue",
		Animations: map[string]string{
			"idle":          "idle.gif",
			"talking":       "talking.gif",
			"romantic_idle": "romantic_idle.gif",
		},
		Dialogs: []Dialog{
			{Trigger: "click", Responses: []string{"Regular hello!"}, Animation: "talking", Cooldown: 5},
		},
		Behavior: Behavior{IdleTimeout: 30, DefaultSize: 128},
		Stats: map[string]StatConfig{
			"affection": {Initial: 0, Max: 100, DegradationRate: 0.1, CriticalThreshold: 10},
		},
		Personality: &PersonalityConfig{
			Traits: map[string]float64{"romanticism": 0.8},
		},
		RomanceDialogs: []DialogExtended{
			{
				Dialog: Dialog{
					Trigger:   "click",
					Responses: []string{"Hi sweetheart! 💕"},
					Animation: "romantic_idle",
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
	
	// Debug: Check if romance features are detected
	t.Logf("HasRomanceFeatures: %v", char.card.HasRomanceFeatures())
	t.Logf("GameState is nil: %v", char.gameState == nil)
	t.Logf("RomanceDialogs count: %d", len(char.card.RomanceDialogs))
	
	if char.gameState != nil {
		t.Logf("Initial affection: %f", char.gameState.GetStat("affection"))
		
		// Set high affection
		char.gameState.Stats["affection"].Current = 25
		t.Logf("After setting affection: %f", char.gameState.GetStat("affection"))
		
		// Test selectRomanceDialog directly
		response := char.selectRomanceDialog("click")
		t.Logf("selectRomanceDialog response: %s", response)
		
		// Test HandleClick
		response = char.HandleClick()
		t.Logf("HandleClick response: %s", response)
	}
}
